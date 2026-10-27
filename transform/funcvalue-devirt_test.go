package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
)

func TestDevirtualizeFuncValues(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/funcvalue-devirt", transform.DevirtualizeFuncValues)
}
