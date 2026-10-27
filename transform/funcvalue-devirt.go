package transform

import (
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// DevirtualizeFuncValues adds a direct call next to calls of func values when
// exactly one function with the same Go signature is used as a func value. A
// pointer comparison picks the direct call and the indirect call stays as the
// fallback. This lets the inliner see through iterators and callbacks that are
// stored in variables or struct fields.
//
// The compiler marks functions used as func values and calls of func values
// with the "tinygo-funcvalue" attribute, which holds the signature.
func DevirtualizeFuncValues(mod llvm.Module) {
	const attrName = "tinygo-funcvalue"

	candidates := map[string][]llvm.Value{}
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		attr := fn.GetStringAttributeAtIndex(-1, attrName)
		if attr.IsNil() {
			continue
		}
		fn.RemoveStringAttributeAtIndex(-1, attrName)
		if fn.IsDeclaration() || !isAddressTaken(fn) {
			continue
		}
		key := attr.GetStringValue()
		candidates[key] = append(candidates[key], fn)
	}

	// Collect calls first, so new direct calls are not visited.
	var calls []llvm.Value
	var targets []llvm.Value
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
				if inst.IsACallInst().IsNil() {
					continue
				}
				attr := inst.GetCallSiteStringAttribute(-1, attrName)
				if attr.IsNil() {
					continue
				}
				callee := inst.CalledValue()
				if !callee.IsAFunction().IsNil() || !callee.IsAInlineAsm().IsNil() {
					continue
				}
				fns := candidates[attr.GetStringValue()]
				if len(fns) != 1 || fns[0].GlobalValueType() != inst.CalledFunctionType() {
					continue
				}
				if inst.InstructionDebugLoc().IsNil() && !fn.Subprogram().IsNil() && !fns[0].Subprogram().IsNil() {
					// The verifier requires a location on inlinable calls with debug info.
					continue
				}
				calls = append(calls, inst)
				targets = append(targets, fns[0])
			}
		}
	}

	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()
	for i, call := range calls {
		target := targets[i]
		block := call.InstructionParent()
		builder.SetInsertPointBefore(call) // use the debug location of this call
		next := llvmutil.SplitBasicBlock(builder, call, block, "funcvalue.next")
		direct := ctx.InsertBasicBlock(next, "funcvalue.direct")
		indirect := ctx.InsertBasicBlock(next, "funcvalue.indirect")

		call.RemoveFromParentAsInstruction()
		builder.SetInsertPointAtEnd(block)
		isTarget := builder.CreateICmp(llvm.IntEQ, call.CalledValue(), target, "funcvalue.is")
		builder.CreateCondBr(isTarget, direct, indirect)

		builder.SetInsertPointAtEnd(indirect)
		builder.Insert(call)
		builder.CreateBr(next)

		builder.SetInsertPointAtEnd(direct)
		args := make([]llvm.Value, call.OperandsCount()-1)
		for j := range args {
			args[j] = call.Operand(j)
		}
		directCall := builder.CreateCall(target.GlobalValueType(), target, args, "")
		directCall.SetInstructionCallConv(call.InstructionCallConv())
		directCall.InstructionSetDebugLoc(call.InstructionDebugLoc())
		builder.CreateBr(next)

		if call.Type().TypeKind() != llvm.VoidTypeKind {
			builder.SetInsertPointBefore(next.FirstInstruction())
			phi := builder.CreatePHI(call.Type(), "")
			call.ReplaceAllUsesWith(phi)
			phi.AddIncoming([]llvm.Value{directCall, call}, []llvm.BasicBlock{direct, indirect})
		}
	}
}

// isAddressTaken returns whether fn is used other than as the callee of a call.
func isAddressTaken(fn llvm.Value) bool {
	for _, use := range getUses(fn) {
		if use.IsACallInst().IsNil() || use.CalledValue() != fn {
			return true
		}
	}
	return false
}
