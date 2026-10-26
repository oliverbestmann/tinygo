target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

declare void @llvm.memcpy.p0.p0.i32(ptr, ptr, i32, i1)
declare void @llvm.memmove.p0.p0.i32(ptr, ptr, i32, i1)
declare void @llvm.memset.p0.i32(ptr, i8, i32, i1)

define void @copyDynamic(ptr %dst, ptr %src, i32 %n) {
  call void @llvm.memcpy.p0.p0.i32(ptr %dst, ptr %src, i32 %n, i1 false)
  ret void
}

define void @moveDynamic(ptr %dst, ptr %src, i32 %n) {
  call void @llvm.memmove.p0.p0.i32(ptr %dst, ptr %src, i32 %n, i1 false)
  ret void
}

define void @clearDynamic(ptr %dst, i32 %n) {
  call void @llvm.memset.p0.i32(ptr %dst, i8 0, i32 %n, i1 false)
  ret void
}

; Constant lengths and volatile copies are left alone.
define void @copyConstant(ptr %dst, ptr %src) {
  call void @llvm.memcpy.p0.p0.i32(ptr %dst, ptr %src, i32 64, i1 false)
  ret void
}

define void @copyVolatile(ptr %dst, ptr %src, i32 %n) {
  call void @llvm.memcpy.p0.p0.i32(ptr %dst, ptr %src, i32 %n, i1 true)
  ret void
}
