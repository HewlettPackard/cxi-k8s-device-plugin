package hpecxi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func hasHPECXI(t *testing.T) bool {
	devices := GetHPECXIs()
	if len(devices) <= 0 {
		return false
	}
	return true
}

func TestCXIDeviceCountConsistent(t *testing.T) {
	if !hasHPECXI(t) {
		t.Skip("Skipping test, no HPE CXI found.")
	}

	devices := GetHPECXIs()

	matches, _ := filepath.Glob("/sys/class/cxi/cxi[0-3]*/device/vendor")

	count := 0
	for _, vidPath := range matches {
		t.Log(vidPath)
		b, err := os.ReadFile(vidPath)
		vid := string(b)

		if err == nil && HPEvendorID == strings.TrimSpace(vid) {
			count++
		} else {
			t.Log(vid)
		}

	}

	if count != len(devices) {
		t.Errorf("NIC counts differ: /sys/module/cxi_core: %d, /sys/class/cxi: %d", len(devices), count)
	}

}

func TestLibs(t *testing.T) {
	libpaths, err := GetLibs()
	if err != nil {
		t.Error(err)
	}
	for _, libpath := range libpaths {
		// Check if the file exists
		if _, err := os.Stat(libpath); os.IsNotExist(err) {
			t.Errorf("File does not exist: %s", libpath)
		} else if err != nil {
			t.Errorf("Error checking file: %v", err)
		} else {
			t.Logf("File exists: %s", libpath)
		}
	}
}
