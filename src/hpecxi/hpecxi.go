package hpecxi

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/golang/glog"
)

const HPEvendorID string = "0x17db"

var LibPaths = map[string]string{
	"libfabric":   "/opt/cray/lib64",
	"libcxi":      "/usr/lib64/libcxi.so",
	"libcxiutils": "/usr/lib64/libcxiutils.so",
}

// GetHPECXIs return a map of HPE Cassini on a node identified by the part of the pci address
func GetHPECXIs() map[string]int {
	if _, err := os.Stat("/sys/module/cxi_core/drivers/"); err != nil {
		glog.Warningf("HPE CXI driver unavailable: %s", err)
		return make(map[string]int)
	}

	matches, _ := filepath.Glob("/sys/module/cxi_core/drivers/pci:cxi_core/[0-9a-fA-F][0-9a-fA-F][0-9a-fA-F][0-9a-fA-F]:*")

	devices := make(map[string]int)

	for _, path := range matches {
		glog.Info(path)
		devPaths, _ := filepath.Glob(path + "/net/*")

		for _, devPath := range devPaths {
			name := filepath.Base(devPath)
			if name[0:3] == "hsn" {
				nic_id, _ := strconv.Atoi(name[len(name)-1:])
				devices[name] = nic_id
			}
		}

	}

	for device, _ := range devices {
		glog.Info("Found device ", device)
	}

	return devices
}

// HPECXI check if a particular card is a HPE CXI NIC by checking the device's vendor ID
func HPECXI(cardName string) bool {
	sysfsVendorPath := "/sys/class/cxi/" + cardName + "/device/vendor"
	b, err := ioutil.ReadFile(sysfsVendorPath)
	if err == nil {
		vid := strings.TrimSpace(string(b))

		if vid == HPEvendorID {
			return true
		}
	} else {
		glog.Errorf("Error opening %s: %s", sysfsVendorPath, err)
	}
	return false
}
