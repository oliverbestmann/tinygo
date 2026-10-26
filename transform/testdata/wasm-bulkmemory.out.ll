target datalayout = "e-m:e-p:32:32-p10:8:8-p20:8:8-i64:64-n32:64-S128-ni:1:10:20"
target triple = "wasm32-unknown-wasi"

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: readwrite)
declare void @llvm.memcpy.p0.p0.i32(ptr noalias nocapture writeonly, ptr noalias nocapture readonly, i32, i1 immarg) #0

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: readwrite)
declare void @llvm.memmove.p0.p0.i32(ptr nocapture writeonly, ptr nocapture readonly, i32, i1 immarg) #0

; Function Attrs: nocallback nofree nounwind willreturn memory(argmem: write)
declare void @llvm.memset.p0.i32(ptr nocapture writeonly, i8, i32, i1 immarg) #1

define void @copyDynamic(ptr %dst, ptr %src, i32 %n) {
  call void @tinygo.memmove.i32(ptr %dst, ptr %src, i32 %n)
  ret void
}

define void @moveDynamic(ptr %dst, ptr %src, i32 %n) {
  call void @tinygo.memmove.i32(ptr %dst, ptr %src, i32 %n)
  ret void
}

define void @clearDynamic(ptr %dst, i32 %n) {
  call void @tinygo.memset.i32(ptr %dst, i8 0, i32 %n)
  ret void
}

define void @copyConstant(ptr %dst, ptr %src) {
  call void @llvm.memcpy.p0.p0.i32(ptr %dst, ptr %src, i32 64, i1 false)
  ret void
}

define void @copyVolatile(ptr %dst, ptr %src, i32 %n) {
  call void @llvm.memcpy.p0.p0.i32(ptr %dst, ptr %src, i32 %n, i1 true)
  ret void
}

define internal void @tinygo.memmove.i32(ptr %0, ptr %1, i32 %2) {
entry:
  %3 = icmp ugt i32 %2, 32
  br i1 %3, label %large, label %small

large:                                            ; preds = %entry
  call void @llvm.memmove.p0.p0.i32(ptr %0, ptr %1, i32 %2, i1 false)
  ret void

small:                                            ; preds = %entry
  %4 = icmp uge i32 %2, 16
  br i1 %4, label %copy, label %small1

copy:                                             ; preds = %small
  %5 = sub i32 %2, 16
  %6 = sub i32 %2, 8
  %7 = getelementptr inbounds i8, ptr %1, i32 0
  %8 = load i64, ptr %7, align 1
  %9 = getelementptr inbounds i8, ptr %1, i32 %5
  %10 = load i64, ptr %9, align 1
  %11 = getelementptr inbounds i8, ptr %1, i32 8
  %12 = load i64, ptr %11, align 1
  %13 = getelementptr inbounds i8, ptr %1, i32 %6
  %14 = load i64, ptr %13, align 1
  %15 = getelementptr inbounds i8, ptr %0, i32 0
  store i64 %8, ptr %15, align 1
  %16 = getelementptr inbounds i8, ptr %0, i32 %5
  store i64 %10, ptr %16, align 1
  %17 = getelementptr inbounds i8, ptr %0, i32 8
  store i64 %12, ptr %17, align 1
  %18 = getelementptr inbounds i8, ptr %0, i32 %6
  store i64 %14, ptr %18, align 1
  ret void

small1:                                           ; preds = %small
  %19 = icmp uge i32 %2, 8
  br i1 %19, label %copy2, label %small3

copy2:                                            ; preds = %small1
  %20 = sub i32 %2, 8
  %21 = getelementptr inbounds i8, ptr %1, i32 0
  %22 = load i64, ptr %21, align 1
  %23 = getelementptr inbounds i8, ptr %1, i32 %20
  %24 = load i64, ptr %23, align 1
  %25 = getelementptr inbounds i8, ptr %0, i32 0
  store i64 %22, ptr %25, align 1
  %26 = getelementptr inbounds i8, ptr %0, i32 %20
  store i64 %24, ptr %26, align 1
  ret void

small3:                                           ; preds = %small1
  %27 = icmp uge i32 %2, 4
  br i1 %27, label %copy4, label %small5

copy4:                                            ; preds = %small3
  %28 = sub i32 %2, 4
  %29 = getelementptr inbounds i8, ptr %1, i32 0
  %30 = load i32, ptr %29, align 1
  %31 = getelementptr inbounds i8, ptr %1, i32 %28
  %32 = load i32, ptr %31, align 1
  %33 = getelementptr inbounds i8, ptr %0, i32 0
  store i32 %30, ptr %33, align 1
  %34 = getelementptr inbounds i8, ptr %0, i32 %28
  store i32 %32, ptr %34, align 1
  ret void

small5:                                           ; preds = %small3
  %35 = icmp uge i32 %2, 2
  br i1 %35, label %copy6, label %small7

copy6:                                            ; preds = %small5
  %36 = sub i32 %2, 2
  %37 = getelementptr inbounds i8, ptr %1, i32 0
  %38 = load i16, ptr %37, align 1
  %39 = getelementptr inbounds i8, ptr %1, i32 %36
  %40 = load i16, ptr %39, align 1
  %41 = getelementptr inbounds i8, ptr %0, i32 0
  store i16 %38, ptr %41, align 1
  %42 = getelementptr inbounds i8, ptr %0, i32 %36
  store i16 %40, ptr %42, align 1
  ret void

small7:                                           ; preds = %small5
  %43 = icmp eq i32 %2, 0
  br i1 %43, label %done, label %copy8

copy8:                                            ; preds = %small7
  %44 = getelementptr inbounds i8, ptr %1, i32 0
  %45 = load i8, ptr %44, align 1
  %46 = getelementptr inbounds i8, ptr %0, i32 0
  store i8 %45, ptr %46, align 1
  ret void

done:                                             ; preds = %small7
  ret void
}

define internal void @tinygo.memset.i32(ptr %0, i8 %1, i32 %2) {
entry:
  %3 = icmp ugt i32 %2, 32
  br i1 %3, label %large, label %small

large:                                            ; preds = %entry
  call void @llvm.memset.p0.i32(ptr %0, i8 %1, i32 %2, i1 false)
  ret void

small:                                            ; preds = %entry
  %4 = zext i8 %1 to i64
  %5 = mul i64 %4, 72340172838076673
  %6 = icmp uge i32 %2, 16
  br i1 %6, label %copy, label %small1

copy:                                             ; preds = %small
  %7 = sub i32 %2, 16
  %8 = sub i32 %2, 8
  %9 = getelementptr inbounds i8, ptr %0, i32 0
  store i64 %5, ptr %9, align 1
  %10 = getelementptr inbounds i8, ptr %0, i32 %7
  store i64 %5, ptr %10, align 1
  %11 = getelementptr inbounds i8, ptr %0, i32 8
  store i64 %5, ptr %11, align 1
  %12 = getelementptr inbounds i8, ptr %0, i32 %8
  store i64 %5, ptr %12, align 1
  ret void

small1:                                           ; preds = %small
  %13 = icmp uge i32 %2, 8
  br i1 %13, label %copy2, label %small3

copy2:                                            ; preds = %small1
  %14 = sub i32 %2, 8
  %15 = getelementptr inbounds i8, ptr %0, i32 0
  store i64 %5, ptr %15, align 1
  %16 = getelementptr inbounds i8, ptr %0, i32 %14
  store i64 %5, ptr %16, align 1
  ret void

small3:                                           ; preds = %small1
  %17 = icmp uge i32 %2, 4
  br i1 %17, label %copy4, label %small5

copy4:                                            ; preds = %small3
  %18 = sub i32 %2, 4
  %19 = trunc i64 %5 to i32
  %20 = getelementptr inbounds i8, ptr %0, i32 0
  store i32 %19, ptr %20, align 1
  %21 = getelementptr inbounds i8, ptr %0, i32 %18
  store i32 %19, ptr %21, align 1
  ret void

small5:                                           ; preds = %small3
  %22 = icmp uge i32 %2, 2
  br i1 %22, label %copy6, label %small7

copy6:                                            ; preds = %small5
  %23 = sub i32 %2, 2
  %24 = trunc i64 %5 to i16
  %25 = getelementptr inbounds i8, ptr %0, i32 0
  store i16 %24, ptr %25, align 1
  %26 = getelementptr inbounds i8, ptr %0, i32 %23
  store i16 %24, ptr %26, align 1
  ret void

small7:                                           ; preds = %small5
  %27 = icmp eq i32 %2, 0
  br i1 %27, label %done, label %copy8

copy8:                                            ; preds = %small7
  %28 = trunc i64 %5 to i8
  %29 = getelementptr inbounds i8, ptr %0, i32 0
  store i8 %28, ptr %29, align 1
  ret void

done:                                             ; preds = %small7
  ret void
}

attributes #0 = { nocallback nofree nounwind willreturn memory(argmem: readwrite) }
attributes #1 = { nocallback nofree nounwind willreturn memory(argmem: write) }
