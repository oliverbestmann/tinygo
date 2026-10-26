package compiler

import (
	"go/types"

	"tinygo.org/x/go-llvm"
)

// getReflectCallStub returns a function that is used by reflect.Value.Call to
// call a func value of the given signature. Arguments and results are passed
// as arrays of pointers to the values. In Go, the stub would look like this:
//
//	func stub(fn *func(...) (...), args, results *unsafe.Pointer)
//
// The stub does nothing after the call, so a panic in the callee propagates to
// the caller of the stub like with any other call.
func (c *compilerContext) getReflectCallStub(sig *types.Signature, typeCodeName string, isLocal bool) llvm.Value {
	if c.funcPtrAddrSpace != 0 {
		// Type code fields are data pointers, so targets with a separate function
		// address space (AVR) get no stub and reflect.Value.Call panics there.
		return llvm.ConstNull(c.dataPtrType)
	}
	name := "reflect/call:" + typeCodeName
	if !isLocal {
		if stub := c.mod.NamedFunction(name); !stub.IsNil() {
			return stub
		}
	}

	stubType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.dataPtrType, c.dataPtrType, c.dataPtrType, c.dataPtrType}, false)
	stub := llvm.AddFunction(c.mod, name, stubType)
	c.addStandardAttributes(stub)
	if isLocal {
		stub.SetLinkage(llvm.InternalLinkage)
	} else {
		stub.SetLinkage(llvm.LinkOnceODRLinkage)
	}
	stub.SetUnnamedAddr(true)

	b := &builder{
		compilerContext: c,
		Builder:         c.ctx.NewBuilder(),
	}
	defer b.Dispose()
	b.SetInsertPointAtEnd(c.ctx.AddBasicBlock(stub, "entry"))

	// The receiver (if any) is not part of a func type.
	sig = types.NewSignatureType(nil, nil, nil, sig.Params(), sig.Results(), sig.Variadic())
	abi := c.getFunctionABI(sig, false)
	fnPtr, context := b.decodeFuncValue(b.CreateLoad(c.getFuncType(sig), stub.Param(0), "fn"))
	loadPtr := func(array llvm.Value, i int) llvm.Value {
		gep := b.CreateInBoundsGEP(c.dataPtrType, array, []llvm.Value{llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false)}, "")
		return b.CreateLoad(c.dataPtrType, gep, "")
	}

	var params []llvm.Value
	var resultPtr llvm.Value
	if abi.indirectResult {
		resultPtr = b.CreateAlloca(abi.resultType, "result")
		params = append(params, resultPtr)
	}
	for i, param := range abi.params {
		ptr := loadPtr(stub.Param(1), i)
		if param.indirect {
			// reflect passes fresh copies of the arguments, so the callee may
			// use this memory directly.
			params = append(params, ptr)
		} else {
			params = append(params, b.CreateLoad(param.llvmType, ptr, ""))
		}
	}
	params = append(params, context)
	ret := b.createCall(c.getLLVMFunctionType(sig), fnPtr, params, "")

	numResults := sig.Results().Len()
	for i := range numResults {
		dst := loadPtr(stub.Param(2), i)
		if abi.indirectResult {
			src := resultPtr
			if numResults > 1 {
				src = b.CreateStructGEP(abi.resultType, resultPtr, i, "")
			}
			b.copyIndirectAggregate(dst, src, c.getLLVMType(sig.Results().At(i).Type()))
			continue
		}
		val := ret
		if numResults > 1 {
			val = b.CreateExtractValue(ret, i, "")
		}
		b.CreateStore(val, dst)
	}
	b.CreateRetVoid()
	return stub
}
