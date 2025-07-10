package hpecxi

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// Mock implementations and helpers for testing

func TestReadPCIFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test_property")
	expected := "test_value"
	if err := os.WriteFile(testFile, []byte(expected), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}
	got := readPCIFile(tmpDir, "test_property")
	if got != expected {
		t.Errorf("readPCIFile() = %q, want %q", got, expected)
	}
}

//TODO: Test for BuildMounts

func TestGetPCISlot(t *testing.T) {
	tmpDir := t.TempDir()
	ueventPath := filepath.Join(tmpDir, "uevent")
	content := "PCI_SLOT_NAME=0000:00:1f.0\n"
	if err := os.WriteFile(ueventPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write uevent file: %v", err)
	}
	got := getPCISlot(ueventPath)
	want := "0000:00:1f.0"
	if got != want {
		t.Errorf("getPCISlot() = %q, want %q", got, want)
	}
}

func TestBuildDevice(t *testing.T) {
	tmpDir := t.TempDir()
	deviceDir := filepath.Join(tmpDir, "device0")
	propertiesDir := filepath.Join(deviceDir, "properties")
	os.MkdirAll(propertiesDir, 0755)

	// Create required files with test content
	os.WriteFile(filepath.Join(deviceDir, "uevent"), []byte("PCI_SLOT_NAME=0000:00:1f.0\n"), 0644)
	os.WriteFile(filepath.Join(deviceDir, "device"), []byte("0x1234"), 0644)
	os.WriteFile(filepath.Join(deviceDir, "local_cpus"), []byte("0-3"), 0644)
	os.WriteFile(filepath.Join(deviceDir, "local_cpulist"), []byte("0,1,2,3"), 0644)
	os.WriteFile(filepath.Join(deviceDir, "numa_node"), []byte("0"), 0644)
	os.WriteFile(filepath.Join(propertiesDir, "cassini_version"), []byte("1.2.3"), 0644)
	os.WriteFile(filepath.Join(propertiesDir, "speed"), []byte("100G"), 0644)

	// Create cxi subdir and a fake device node
	cxiDir := filepath.Join(deviceDir, "cxi")
	os.MkdirAll(cxiDir, 0755)
	os.WriteFile(filepath.Join(cxiDir, "cxi0"), []byte{}, 0644)

	got := BuildDevice(deviceDir, "device0")

	if got.Name != "device0" {
		t.Errorf("BuildDevice.Name = %q, want %q", got.Name, "device0")
	}
	if got.PCIAddress != "0000:00:1f.0" {
		t.Errorf("BuildDevice.PCIAddress = %q, want %q", got.PCIAddress, "0000:00:1f.0")
	}
	if got.DeviceId != 0 {
		t.Errorf("BuildDevice.DeviceId = %d, want %d", got.DeviceId, 0)
	}
	if got.LocalCPUs != "0-3" {
		t.Errorf("BuildDevice.LocalCPUs = %q, want %q", got.LocalCPUs, "0-3")
	}
	if got.LocalCPUList != "0,1,2,3" {
		t.Errorf("BuildDevice.LocalCPUList = %q, want %q", got.LocalCPUList, "0,1,2,3")
	}
	if got.NumaNode != "0" {
		t.Errorf("BuildDevice.NumaNode = %q, want %q", got.NumaNode, "0")
	}
	if got.Version != "1.2.3" {
		t.Errorf("BuildDevice.Version = %q, want %q", got.Version, "1.2.3")
	}
	if got.Speed != "100G" {
		t.Errorf("BuildDevice.Speed = %q, want %q", got.Speed, "100G")
	}
}

func TestReadDeviceNumber(t *testing.T) {
	tmpDir := t.TempDir()
	cxiDir := filepath.Join(tmpDir, "cxi")
	if err := os.MkdirAll(cxiDir, 0755); err != nil {
		t.Fatalf("Failed to create cxi directory: %v", err)
	}

	// Create fake cxi device files: cxi0, cxi1
	for i := 0; i < 2; i++ {
		fname := filepath.Join(cxiDir, "cxi"+strconv.Itoa(i))
		if err := os.WriteFile(fname, []byte{}, 0644); err != nil {
			t.Fatalf("Failed to create fake cxi device file: %v", err)
		}
	}

	id, err := readDeviceNumber(tmpDir)
	if err != nil {
		t.Fatalf("readDeviceNumber returned error: %v", err)
	}
	if id != 0 {
		t.Errorf("readDeviceNumber() = %d, want 0", id)
	}

	// Test error case: no cxi devices
	emptyDir := t.TempDir()
	_, err = readDeviceNumber(emptyDir)
	if err == nil {
		t.Errorf("readDeviceNumber() with no cxi devices: expected error, got nil")
	}
}

func TestDiscoverEnvVars(t *testing.T) {
	gotEnv := DiscoverEnvVars("../../test/data/envVars.yaml")

	expectedEnv := []string{
		"VAR1=\"value1\"",
		"VAR2=\"value2\"",
		"VAR3=3",
	}

	if len(gotEnv) != len(expectedEnv) {
		t.Fatalf("Expected %d env variables, got %d", len(expectedEnv), len(gotEnv))
	}

	for i, envVar := range expectedEnv {
		if gotEnv[i] != envVar {
			t.Errorf("Expected %s, got %s", envVar, gotEnv[i])
		}
	}
}
