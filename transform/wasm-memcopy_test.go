package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
)

func TestOptimizeWasmBulkMemory(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/wasm-bulkmemory", transform.OptimizeWasmBulkMemory)
}
