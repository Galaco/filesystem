package filesystem

import (
	"reflect"
	"testing"

	"github.com/galaco/vpk2"
)

func TestGetFile(t *testing.T) {
	t.Skip()
}

func TestRegisterPakfile(t *testing.T) {
	t.Skip()
}

func TestRegisterVpk(t *testing.T) {
	fs := NewFileSystem()
	fs.RegisterVpk("b", &vpk.VPK{})
	fs.RegisterVpk("a", &vpk.VPK{})
	fs.RegisterVpk("c", &vpk.VPK{})
	// Registering again keeps the original search order
	fs.RegisterVpk("a", &vpk.VPK{})

	expected := []string{"b", "a", "c"}
	actual := fs.EnumerateResourcePaths()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected vpks to be searched in order %v, got %v", expected, actual)
	}
}

func TestUnregisterVpk(t *testing.T) {
	fs := NewFileSystem()
	fs.RegisterVpk("a", &vpk.VPK{})
	fs.RegisterVpk("b", &vpk.VPK{})
	fs.RegisterVpk("c", &vpk.VPK{})
	fs.UnregisterVpk("b")

	expected := []string{"a", "c"}
	actual := fs.EnumerateResourcePaths()
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("expected remaining vpks %v, got %v", expected, actual)
	}
}

func TestRegisterLocalDirectory(t *testing.T) {
	fs := NewFileSystem()
	dir := "foo/bar/baz"
	fs.RegisterLocalDirectory(dir)
	found := false
	for _, path := range fs.EnumerateResourcePaths() {
		if path == dir {
			found = true
			break
		}
	}
	if found == false {
		t.Error("local filepath was not found in registered paths")
	}
}

func TestUnregisterLocalDirectory(t *testing.T) {
	fs := NewFileSystem()
	dir := "foo/bar/baz"
	fs.RegisterLocalDirectory(dir)
	fs.UnregisterLocalDirectory(dir)
	found := false
	for _, path := range fs.EnumerateResourcePaths() {
		if path == dir {
			found = true
			break
		}
	}
	if found == true {
		t.Error("local filepath was not found in registered paths")
	}
}

func TestUnregisterPakfile(t *testing.T) {
	t.Skip()
}
