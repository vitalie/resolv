package resolv_test

import (
	"testing"
)

func TestTrue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
}
