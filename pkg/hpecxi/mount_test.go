package hpecxi

import (
	"reflect"
	"testing"
)

func TestMountInfoClone(t *testing.T) {
	orig := &MountInfo{
		Name:          "libfabric",
		HostPath:      "/foo/bar",
		ContainerPath: "/foo/bar",
		Options:       []string{"ro", "nosuid"},
		Type:          "d",
	}
	clone := orig.Clone()

	if !reflect.DeepEqual(orig, clone) {
		t.Errorf("Clone() = %+v, want %+v", clone, orig)
	}
	if orig == clone {
		t.Errorf("Clone() returned the same pointer, want a new one")
	}
}

func TestMountsInfoClone(t *testing.T) {
	orig := MountsInfo{
		"bar": &MountInfo{
			Name:          "libfabric",
			HostPath:      "/foo/bar",
			ContainerPath: "/foo/bar",
			Options:       []string{"ro", "nosuid"},
			Type:          "d",
		},
		"baz": &MountInfo{
			Name:          "libcxi",
			HostPath:      "/foo/baz",
			ContainerPath: "/foo/baz",
			Options:       []string{"ro"},
			Type:          "d",
		},
	}
	clone := orig.Clone()

	if !reflect.DeepEqual(orig, clone) {
		t.Errorf("MountsInfo.Clone() = %+v, want %+v", clone, orig)
	}
	if &orig == &clone {
		t.Errorf("MountsInfo.Clone() returned the same pointer, want a new one")
	}
	for k := range orig {
		if orig[k] == clone[k] {
			t.Errorf("MountInfo for key %q: Clone() returned the same pointer, want a new one", k)
		}
	}
}
