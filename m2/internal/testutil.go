package internal

import (
	"os"
	"testing"
)

func TouchFile(t *testing.T, path string) {
	t.Helper()
	if _, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644); err != nil {
		t.Fatalf("failed to create file '%s': %v", path, err)
	}
}
