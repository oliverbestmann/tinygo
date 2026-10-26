package transform

import (
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

// maxInlineMemCopy is the largest length that the helpers created by
// OptimizeWasmBulkMemory copy or fill with plain loads and stores.
const maxInlineMemCopy = 32

// OptimizeWasmBulkMemory replaces llvm.memcpy, llvm.memmove and llvm.memset
// calls that have a non-constant length with calls to a helper. The helper
// handles lengths up to maxInlineMemCopy bytes with plain loads and stores and
// only uses memory.copy or memory.fill for longer lengths. Browsers run these
// bulk memory instructions as calls into the host, which is slow for the small
// sizes that Go code often copies, like component values of a runtime size.
func OptimizeWasmBulkMemory(mod llvm.Module) {
	builder := mod.Context().NewBuilder()
	defer builder.Dispose()

	// Collect the calls first, so the calls inside the helpers stay as they are.
	var calls []llvm.Value
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		name := fn.Name()
		if !strings.HasPrefix(name, "llvm.memset.") && !strings.HasPrefix(name, "llvm.memcpy.") && !strings.HasPrefix(name, "llvm.memmove.") {
			continue
		}
		for _, call := range getUses(fn) {
			if call.IsACallInst().IsNil() || call.CalledValue() != fn {
				continue
			}
			isVolatile := call.Operand(3)
			if call.Operand(2).IsAConstantInt().IsNil() && !isVolatile.IsAConstantInt().IsNil() && isVolatile.ZExtValue() == 0 {
				calls = append(calls, call)
			}
		}
	}

	helpers := map[string]llvm.Value{}
	for _, call := range calls {
		length := call.Operand(2)
		isSet := strings.HasPrefix(call.CalledValue().Name(), "llvm.memset.")
		kind := "memmove"
		if isSet {
			kind = "memset"
		}
		helperName := "tinygo." + kind + ".i" + strconv.Itoa(length.Type().IntTypeWidth())
		helper := helpers[helperName]
		if helper.IsNil() {
			helper = createBulkMemoryHelper(mod, builder, helperName, isSet, call.Operand(1).Type(), length.Type())
			helpers[helperName] = helper
		}

		builder.SetInsertPointBefore(call)
		newCall := builder.CreateCall(helper.GlobalValueType(), helper, []llvm.Value{call.Operand(0), call.Operand(1), length}, "")
		newCall.InstructionSetDebugLoc(call.InstructionDebugLoc())
		call.EraseFromParentAsInstruction()
	}
}

// createBulkMemoryHelper creates a function like memmove(dst, src, n) or
// memset(dst, value, n). For n of at least size (16, 8, 4, 2 or 1) it does the
// first and the last size bytes at once, which may overlap. All loads happen
// before the stores, so a memmove with overlapping buffers is correct.
func createBulkMemoryHelper(mod llvm.Module, builder llvm.Builder, name string, isSet bool, srcType, lengthType llvm.Type) llvm.Value {
	ctx := mod.Context()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	fnType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType, srcType, lengthType}, false)
	fn := llvm.AddFunction(mod, name, fnType)
	fn.SetLinkage(llvm.InternalLinkage)
	dst, src, n := fn.Param(0), fn.Param(1), fn.Param(2)

	lengthSuffix := ".i" + strconv.Itoa(lengthType.IntTypeWidth())
	intrinsic := "llvm.memmove.p0.p0" + lengthSuffix
	intrinsicParams := []llvm.Type{ptrType, ptrType, lengthType, ctx.Int1Type()}
	if isSet {
		intrinsic = "llvm.memset.p0" + lengthSuffix
		intrinsicParams[1] = ctx.Int8Type()
	}
	intrinsicFn := mod.NamedFunction(intrinsic)
	if intrinsicFn.IsNil() {
		intrinsicFn = llvm.AddFunction(mod, intrinsic, llvm.FunctionType(ctx.VoidType(), intrinsicParams, false))
	}

	entry := ctx.AddBasicBlock(fn, "entry")
	large := ctx.AddBasicBlock(fn, "large")
	builder.SetInsertPointAtEnd(entry)
	isLarge := builder.CreateICmp(llvm.IntUGT, n, llvm.ConstInt(lengthType, maxInlineMemCopy, false), "")
	next := ctx.AddBasicBlock(fn, "small")
	builder.CreateCondBr(isLarge, large, next)

	builder.SetInsertPointAtEnd(large)
	builder.CreateCall(intrinsicFn.GlobalValueType(), intrinsicFn, []llvm.Value{dst, src, n, llvm.ConstInt(ctx.Int1Type(), 0, false)}, "")
	builder.CreateRetVoid()

	var splat llvm.Value
	if isSet {
		builder.SetInsertPointAtEnd(next)
		splat = builder.CreateMul(builder.CreateZExt(src, ctx.Int64Type(), ""), llvm.ConstInt(ctx.Int64Type(), 0x0101010101010101, false), "")
	}

	for _, size := range []uint64{16, 8, 4, 2, 1} {
		builder.SetInsertPointAtEnd(next)
		copyBlock := ctx.AddBasicBlock(fn, "copy")
		if size == 1 {
			next = ctx.AddBasicBlock(fn, "done")
			builder.CreateCondBr(builder.CreateICmp(llvm.IntEQ, n, llvm.ConstInt(lengthType, 0, false), ""), next, copyBlock)
		} else {
			next = ctx.AddBasicBlock(fn, "small")
			builder.CreateCondBr(builder.CreateICmp(llvm.IntUGE, n, llvm.ConstInt(lengthType, size, false), ""), copyBlock, next)
		}

		builder.SetInsertPointAtEnd(copyBlock)
		chunk := min(size, 8)
		chunkType := ctx.IntType(int(chunk) * 8)
		var offsets []llvm.Value
		for off := uint64(0); off < size; off += chunk {
			offsets = append(offsets, llvm.ConstInt(lengthType, off, false))
			if size > 1 {
				tail := builder.CreateSub(n, llvm.ConstInt(lengthType, size-off, false), "")
				offsets = append(offsets, tail)
			}
		}
		gep := func(base, offset llvm.Value) llvm.Value {
			return builder.CreateInBoundsGEP(ctx.Int8Type(), base, []llvm.Value{offset}, "")
		}
		values := make([]llvm.Value, len(offsets))
		var value llvm.Value
		if isSet {
			value = builder.CreateTrunc(splat, chunkType, "")
		}
		for i, off := range offsets {
			if isSet {
				values[i] = value
				continue
			}
			values[i] = builder.CreateLoad(chunkType, gep(src, off), "")
			values[i].SetAlignment(1)
		}
		for i, off := range offsets {
			builder.CreateStore(values[i], gep(dst, off)).SetAlignment(1)
		}
		builder.CreateRetVoid()
	}

	builder.SetInsertPointAtEnd(next)
	builder.CreateRetVoid()
	return fn
}
