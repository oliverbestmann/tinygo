target datalayout = "e-m:e-p:32:32-i64:64-n32:64-S128"
target triple = "wasm32-unknown-wasi"

@seq = global ptr @main.all$1
@cb1 = global ptr @main.callback1
@cb2 = global ptr @main.callback2
@pred = global ptr @main.pred

; Only func value with signature "iter".
define internal void @main.all$1(ptr %yield.context, ptr %yield.funcptr, ptr %context) #0 {
entry:
  %ok = call i1 %yield.funcptr(i32 1, ptr %yield.context) #2
  ret void
}

; Two func values with signature "yield", so calls of it are left alone.
define internal i1 @main.callback1(i32 %x, ptr %context) #1 {
entry:
  ret i1 true
}

define internal i1 @main.callback2(i32 %x, ptr %context) #1 {
entry:
  ret i1 false
}

define internal i1 @main.pred(i32 %x, ptr %context) #6 {
entry:
  %ok = icmp eq i32 %x, 5
  ret i1 %ok
}

; Tagged, but its address is not taken.
define internal void @main.unused(ptr %context) #3 {
entry:
  ret void
}

define void @main.callIter(ptr %fn, ptr %fnctx, ptr %yieldctx, ptr %yieldfn) {
entry:
  call void %fn(ptr %yieldctx, ptr %yieldfn, ptr %fnctx) #4
  ret void
}

define i1 @main.callResult(ptr %fn, ptr %fnctx) {
entry:
  %result = call i1 %fn(i32 5, ptr %fnctx) #7
  %x = xor i1 %result, true
  ret i1 %x
}

define void @main.callUnused(ptr %fn) {
entry:
  call void %fn(ptr undef) #5
  call void @main.unused(ptr undef)
  ret void
}

attributes #0 = { "tinygo-funcvalue"="iter" }
attributes #1 = { "tinygo-funcvalue"="yield" }
attributes #2 = { "tinygo-funcvalue"="yield" }
attributes #3 = { "tinygo-funcvalue"="unused" }
attributes #4 = { "tinygo-funcvalue"="iter" }
attributes #5 = { "tinygo-funcvalue"="unused" }
attributes #6 = { "tinygo-funcvalue"="pred" }
attributes #7 = { "tinygo-funcvalue"="pred" }
