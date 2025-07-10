package internal

import (
	"os"
	"testing"
)

type TestDir struct {
	Path string

	t *testing.T
}

func NewTestDir(t *testing.T) *TestDir {
	t.Helper()
	return &TestDir{
		Path: t.TempDir(),
		t:    t,
	}
}

func (d *TestDir) TouchFile(name string) string {
	d.t.Helper()
	path := d.Path + "/" + name
	if _, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0644); err != nil {
		d.t.Fatalf("failed to create file '%s': %v", path, err)
	}
	return path
}

func (d *TestDir) WriteFile(name, content string, mode os.FileMode) string {
	d.t.Helper()
	path := d.Path + "/" + name
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		d.t.Fatalf("failed to write file '%s': %v", path, err)
	}
	return path
}
