//go:build tinygo.fatalruntimeerrors

package runtime

// Runtime errors abort instead of unwinding. With Asyncify this avoids
// instrumenting every function that contains a bounds or nil check.
const fatalRuntimeErrors = true
