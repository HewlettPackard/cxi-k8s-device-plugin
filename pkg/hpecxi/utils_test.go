package hpecxi

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestGetSysfsRoot_Default(t *testing.T) {
	os.Unsetenv(SysfsEnvVarName)
	got := GetSysfsRoot("somepath")
	want := "/sys"
	if got != want {
		t.Errorf("GetSysfsRoot() = %s; want %s", got, want)
	}
}

func TestGetSysfsRoot_Custom(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := "somepath"
	fullPath := path.Join(tmpDir, testPath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		t.Fatalf("Failed to create test sysfs dir: %v", err)
	}
	os.Setenv(SysfsEnvVarName, tmpDir)
	defer os.Unsetenv(SysfsEnvVarName)
	got := GetSysfsRoot(testPath)
	if got != tmpDir {
		t.Errorf("GetSysfsRoot() = %s; want %s", got, tmpDir)
	}
}

func TestGetDevRoot_Default(t *testing.T) {
	os.Unsetenv(DevfsEnvVarName)
	got := GetDevRoot("somepath")
	want := devfsDefaultRoot
	if got != want {
		t.Errorf("GetDevRoot() = %s; want %s", got, want)
	}
}

func TestGetDevRoot_Custom(t *testing.T) {
	tmpDir := t.TempDir()
	testPath := "somepath"
	fullPath := path.Join(tmpDir, testPath)
	if err := os.MkdirAll(fullPath, 0755); err != nil {
		t.Fatalf("Failed to create test devfs dir: %v", err)
	}
	os.Setenv(DevfsEnvVarName, tmpDir)
	defer os.Unsetenv(DevfsEnvVarName)
	got := GetDevRoot(testPath)
	if got != tmpDir {
		t.Errorf("GetDevRoot() = %s; want %s", got, tmpDir)
	}
}

func TestPciInfoFromDeviceUID(t *testing.T) {
	uid := "0000-00-01-0-0x0000"
	wantPCI := "0000:00:01.0"
	wantID := "0x0000"
	gotPCI, gotID := PciInfoFromDeviceUID(uid)
	if gotPCI != wantPCI || gotID != wantID {
		t.Errorf("PciInfoFromDeviceUID(%q) = (%q, %q); want (%q, %q)", uid, gotPCI, gotID, wantPCI, wantID)
	}
}

func TestDeviceUIDFromPCIinfo(t *testing.T) {
	pci := "0000:00:01.0"
	id := "0x0000"
	want := "0000-00-01-0-0x0000"
	got := DeviceUIDFromPCIinfo(pci, id)
	if got != want {
		t.Errorf("DeviceUIDFromPCIinfo(%q, %q) = %q; want %q", pci, id, got, want)
	}
}

func TestGetLibfabricRoot_CustomFound(t *testing.T) {
	tmpDir := t.TempDir()
	libPath := filepath.Join(tmpDir, "libfabric.so")
	os.WriteFile(libPath, []byte{}, 0644)
	os.Setenv(LibfabricEnvVarName, tmpDir)
	defer os.Unsetenv(LibfabricEnvVarName)
	got, err := GetLibfabricRoot()
	if err != nil {
		t.Fatalf("GetLibfabricRoot() returned error: %v", err)
	}
	if got != tmpDir {
		t.Errorf("GetLibfabricRoot() = %s; want %s", got, tmpDir)
	}
}

func TestGetLibcxiRoot_CustomFound(t *testing.T) {
	tmpDir := t.TempDir()
	libPath := filepath.Join(tmpDir, "libcxi.so")
	os.WriteFile(libPath, []byte{}, 0644)
	os.Setenv(LibcxiEnvVarName, tmpDir)
	defer os.Unsetenv(LibcxiEnvVarName)
	got, err := GetLibcxiRoot()
	if err != nil {
		t.Fatalf("GetLibcxiRoot() returned error: %v", err)
	}
	if got != tmpDir {
		t.Errorf("GetLibcxiRoot() = %s; want %s", got, tmpDir)
	}
}
