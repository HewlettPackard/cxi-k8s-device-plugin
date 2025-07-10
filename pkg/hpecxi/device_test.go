package hpecxi

import (
	"os"
	"path/filepath"
	"testing"
)

// Test Clone method of DeviceInfo
func TestDeviceInfoClone(t *testing.T) {
	orig := &DeviceInfo{
		Name:         "cxi0",
		UID:          "0000-00-00-0-0x0000",
		DeviceId:     0,
		PCIAddress:   "0000:00:00.0",
		LocalCPUs:    "ffff0000",
		LocalCPUList: "0-3",
		NumaNode:     "0",
		Version:      "1.1",
		Speed:        "200000",
	}
	clone := orig.Clone()
	if clone == orig {
		t.Error("Clone should return a new pointer")
	}
	if *clone != *orig {
		t.Error("Cloned DeviceInfo does not match original")
	}
}

// Test Clone method of DevicesInfo
func TestDevicesInfoClone(t *testing.T) {
	orig := DevicesInfo{
		"0000-00-00-0-0x0000": &DeviceInfo{Name: "cxi0", UID: "0000-00-00-0-0x0000"},
		"0000-00-00-0-0x0001": &DeviceInfo{Name: "cxi1", UID: "0000-00-00-0-0x0001"},
	}
	clone := orig.Clone()
	if &clone == &orig {
		t.Error("Clone should return a new map")
	}
	for k, v := range orig {
		if clone[k] == v {
			t.Errorf("Clone should return new pointers for DeviceInfo, got same pointer for key %s", k)
		}
		if *clone[k] != *v {
			t.Errorf("Cloned DeviceInfo does not match original for key %s", k)
		}
	}
}

// Test GetDevPath returns the correct path
func TestGetDevPath(t *testing.T) {
	expected := filepath.Join(GetDevRoot(DevfsPath), DevfsPath)
	got := GetDevPath()
	if got != expected {
		t.Errorf("GetDevPath() = %s; want %s", got, expected)
	}
}

// Test HPECXI returns true for correct vendor ID and false otherwise
func TestHPECXI(t *testing.T) {
	tmpDir := t.TempDir()
	fakeUID := "fakeuid"
	sysfsVendorPath := filepath.Join(tmpDir, fakeUID, "/vendor")
	err := os.MkdirAll(filepath.Dir(sysfsVendorPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create sysfs vendor dir: %v", err)
	}
	err = os.WriteFile(sysfsVendorPath, []byte(HPEvendorID), 0644)
	if err != nil {
		t.Fatalf("Failed to write vendor file: %v", err)
	}
	device := DeviceInfo{UID: fakeUID}
	if !HPECXI(device, tmpDir) {
		t.Error("HPECXI should return true for correct vendor ID")
	}
	// Write wrong vendor ID
	err = os.WriteFile(sysfsVendorPath, []byte("0x1234"), 0644)
	if err != nil {
		t.Fatalf("Failed to write vendor file: %v", err)
	}
	if HPECXI(device, tmpDir) {
		t.Error("HPECXI should return false for incorrect vendor ID")
	}
}
