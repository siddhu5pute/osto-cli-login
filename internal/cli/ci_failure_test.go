package cli

import "testing"

func TestCIFailure(t *testing.T) {
	t.Fatal("intentional CI failure")
}
