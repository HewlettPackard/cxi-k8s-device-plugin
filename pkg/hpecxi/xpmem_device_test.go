package hpecxi

import (
	"path/filepath"
	"testing"
)

func TestGetXpmemDevPath(t *testing.T) {
	expected := filepath.Join(GetDevRoot(XpmemDevfsPath), XpmemDevfsPath)
	got := GetXpmemDevPath()
	if got != expected {
		t.Errorf("GetXpmemDevPath() = %s; want %s", got, expected)
	}
}
