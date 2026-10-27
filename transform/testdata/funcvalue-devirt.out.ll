target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-wasi"

@seq = global ptr @"main.all$1"
@cb1 = global ptr @main.callback1
@cb2 = global ptr @main.callback2
@pred = global ptr @main.pred

define internal void @"main.all$1"(ptr %yield.context, ptr %yield.funcptr, ptr %context) {
entry:
  %ok = call i1 %yield.funcptr(i32 1, ptr %yield.context) #0
  ret void
}

define internal i1 @main.callback1(i32 %x, ptr %context) {
entry:
  ret i1 true
}

define internal i1 @main.callback2(i32 %x, ptr %context) {
entry:
  ret i1 false
}

define internal i1 @main.pred(i32 %x, ptr %context) {
entry:
  %ok = icmp eq i32 %x, 5
  ret i1 %ok
}

define internal void @main.unused(ptr %context) {
entry:
  ret void
}

define void @main.callIter(ptr %fn, ptr %fnctx, ptr %yieldctx, ptr %yieldfn) {
entry:
  %funcvalue.is = icmp eq ptr %fn, @"main.all$1"
  br i1 %funcvalue.is, label %funcvalue.direct, label %funcvalue.indirect

funcvalue.direct:                                 ; preds = %entry
  call void @"main.all$1"(ptr %yieldctx, ptr %yieldfn, ptr %fnctx)
  br label %funcvalue.next

funcvalue.indirect:                               ; preds = %entry
  call void %fn(ptr %yieldctx, ptr %yieldfn, ptr %fnctx) #1
  br label %funcvalue.next

funcvalue.next:                                   ; preds = %funcvalue.direct, %funcvalue.indirect
  ret void
}

define i1 @main.callResult(ptr %fn, ptr %fnctx) {
entry:
  %funcvalue.is = icmp eq ptr %fn, @main.pred
  br i1 %funcvalue.is, label %funcvalue.direct, label %funcvalue.indirect

funcvalue.direct:                                 ; preds = %entry
  %0 = call i1 @main.pred(i32 5, ptr %fnctx)
  br label %funcvalue.next

funcvalue.indirect:                               ; preds = %entry
  %1 = call i1 %fn(i32 5, ptr %fnctx) #2
  br label %funcvalue.next

funcvalue.next:                                   ; preds = %funcvalue.direct, %funcvalue.indirect
  %2 = phi i1 [ %0, %funcvalue.direct ], [ %1, %funcvalue.indirect ]
  %3 = xor i1 %2, true
  ret i1 %3
}

define void @main.callUnused(ptr %fn) {
entry:
  call void %fn(ptr undef) #3
  call void @main.unused(ptr undef)
  ret void
}

attributes #0 = { "tinygo-funcvalue"="yield" }
attributes #1 = { "tinygo-funcvalue"="iter" }
attributes #2 = { "tinygo-funcvalue"="pred" }
attributes #3 = { "tinygo-funcvalue"="unused" }
