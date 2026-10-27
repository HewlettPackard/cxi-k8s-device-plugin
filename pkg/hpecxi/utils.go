package hpecxi

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

const (
	// SysfsEnvVarName is the environment variable name for overriding the default sysfs root path
	SysfsEnvVarName = "SYSFS_ROOT"
	// sysfsDefaultRoot is the default sysfs root directory on Linux systems
	sysfsDefaultRoot = "/sys"

	// DevfsEnvVarName is the environment variable name for overriding the default device filesystem root
	DevfsEnvVarName = "DEVFS_ROOT"
	// devfsDefaultRoot is the default device filesystem root directory
	devfsDefaultRoot = "/dev"

	// LibfabricEnvVarName is the environment variable name for specifying libfabric library location
	LibfabricEnvVarName = "OFI_ROOT"
	// libfabricDefaultRoot is the default path where libfabric libraries are typically installed
	libfabricDefaultRoot = "/opt/cray/lib64"
	// libfabricName is the base name of the libfabric shared library
	libfabricName = "libfabric.so"

	// LibcxiEnvVarName is the environment variable name for specifying libcxi library location
	LibcxiEnvVarName = "CXI_ROOT"
	// libcxiDefaultRoot is the default path where libcxi libraries are typically installed
	libcxiDefaultRoot = "/usr/lib64"
	// libcxiName is the base name of the libcxi shared library
	libcxiName = "libcxi.so"

	// PCIAddressLength defines the expected length of a PCI address in DBDF format (0000:00:00.0)
	PCIAddressLength = len("0000:00:00.0")

	// virtualDevicesEnvVarName is the environment variable name for configuring virtual devices per physical device
	virtualDevicesEnvVarName = "CXI_VIRTUAL_DEVICES"
	// virtualDevicesDefaultValue is the default number of virtual devices per physical device
	virtualDevicesDefaultValue = "0"
)

// GetSysfsRoot returns the sysfs root directory to use for device discovery.
// It first checks the SYSFS_ROOT environment variable for a custom location,
// validating that the specified path exists. If not found or invalid, it falls back to the default "/sys".
func GetSysfsRoot(sysfsPath string) string {
	sysfsRoot, found := os.LookupEnv(SysfsEnvVarName)

	if found {
		// Validate that the custom sysfs path exists
		if _, err := os.Stat(path.Join(sysfsRoot, sysfsPath)); err == nil {
			klog.V(4).Infof("using custom sysfs location: %v\n", sysfsRoot)
			return sysfsRoot
		} else {
			klog.V(4).Infof("could not find sysfs at '%v' from %v env var: %v\n", sysfsPath, SysfsEnvVarName, err)
		}
	}

	klog.V(4).Infof("using default sysfs location: %v\n", sysfsDefaultRoot)
	return sysfsDefaultRoot
}

// GetDevRoot returns the device filesystem root directory to use for device access.
// It first checks the DEVFS_ROOT environment variable for a custom location,
// validating that the specified path exists. If not found or invalid, it falls back to the default "/dev".
func GetDevRoot(devPath string) string {
	devfsRoot, found := os.LookupEnv(DevfsEnvVarName)

	if found {
		// Validate that the custom devfs path exists
		if _, err := os.Stat(path.Join(devfsRoot, devPath)); err == nil {
			klog.V(4).Infof("using custom devfs location: %v\n", devfsRoot)
			return devfsRoot
		} else {
			klog.V(4).Infof("could not find devfs at '%v' from %v env var: %v\n", devPath, DevfsEnvVarName, err)
		}
	}

	klog.V(4).Infof("using default devfs root: %v\n", devfsDefaultRoot)
	return devfsDefaultRoot
}

// GetLibfabricRoot locates the libfabric library installation directory.
// It first checks the OFI_ROOT environment variable for a custom location,
// then falls back to the default path. Returns an error if libfabric is not found in either location.
func GetLibfabricRoot() (string, error) {
	libfabricRoot, found := os.LookupEnv(LibfabricEnvVarName)
	if found {
		// Check if libfabric exists in the custom location
		exists := false
		err := existInPath(libfabricName, libfabricRoot, &exists)
		if err != nil {
			return "", err
		}
		if exists {
			klog.V(4).Infof("using custom Libfabric location: %v\n", libfabricRoot)
			return libfabricRoot, nil
		}
	}
	// Fall back to default location
	exists := false
	err := existInPath(libfabricName, libfabricDefaultRoot, &exists)
	if err != nil {
		return "", err
	}
	if exists {
		klog.V(4).Infof("using default Libfabric root: %v\n", libfabricDefaultRoot)
		return libfabricDefaultRoot, nil
	}
	return "", fmt.Errorf("no Libfabric found")
}

// GetLibcxiRoot locates the libcxi library installation directory.
// It first checks the CXI_ROOT environment variable for a custom location,
// then falls back to the default path. Returns an error if libcxi is not found in either location.
func GetLibcxiRoot() (string, error) {
	libcxiRoot, found := os.LookupEnv(LibcxiEnvVarName)
	if found {
		// Check if libcxi exists in the custom location
		exists := false
		err := existInPath(libcxiName, libcxiRoot, &exists)
		if err != nil {
			return "", err
		}
		if exists {
			klog.V(4).Infof("using custom libcxi location: %v\n", libcxiRoot)
			return libcxiRoot, nil
		}
	}
	// Fall back to default location
	exists := false
	err := existInPath(libcxiName, libcxiDefaultRoot, &exists)
	if err != nil {
		return "", err
	}
	if exists {
		klog.V(4).Infof("using default libcxi root: %v\n", libcxiDefaultRoot)
		return libcxiDefaultRoot, nil
	}
	return "", fmt.Errorf("no libcxi found")
}

// existInPath checks if a file with the given name prefix exists in the specified directory.
// It sets the exists boolean pointer to true if a matching file is found, false otherwise.
// Returns an error if the directory cannot be read.
func existInPath(libName, libPath string, exists *bool) error {
	fileInfos, err := os.ReadDir(libPath)
	if err != nil {
		klog.Errorf("Error checking for %s in %s: %v", libName, libPath, err)
		return err
	}
	*exists = false
	// Check each file to see if it starts with the library name
	for _, fileInfo := range fileInfos {
		if strings.HasPrefix(fileInfo.Name(), libName) {
			*exists = true
		}
	}
	return nil
}

// GetVirtualDevicesCount reads the number of virtual devices per physical device from environment.
// It checks the CXI_VIRTUAL_DEVICES environment variable, defaulting to 0 if not set or invalid.
// Returns 0 if the environment variable contains an invalid integer value.
func GetVirtualDevicesCount() int {
	virtualDevicesPerPhysical, found := os.LookupEnv(virtualDevicesEnvVarName)
	if !found {
		virtualDevicesPerPhysical = virtualDevicesDefaultValue
	}
	count, err := strconv.Atoi(virtualDevicesPerPhysical)
	if err != nil {
		klog.Errorf("Error parsing %s: %v", virtualDevicesEnvVarName, err)
		return 0
	}
	return count
}

// PciInfoFromDeviceUID extracts PCI address and device ID from a device UID.
// Converts from RFC1123-compatible format (0000-00-01-0-0x0000) to standard PCI format (0000:00:01.0).
// Returns the PCI address in DBDF notation and the device ID separately.
func PciInfoFromDeviceUID(deviceUID string) (string, string) {
	// 0000-00-01-0-0x0000 -> 0000:00:01.0, 0x0000
	rfc1123PCIaddress := deviceUID[:PCIAddressLength]
	pciAddress := strings.Replace(strings.Replace(rfc1123PCIaddress, "-", ":", 2), "-", ".", 1)
	deviceId := deviceUID[PCIAddressLength+1:]

	return pciAddress, deviceId
}

// DeviceUIDFromPCIinfo creates a device UID from PCI address and device ID.
// Converts from standard PCI format (0000:00:01.0) to RFC1123-compatible format (0000-00-01-0-0x0000).
// This format is safe for use as Kubernetes resource names and identifiers.
func DeviceUIDFromPCIinfo(pciAddress string, pciid string) string {
	// 0000:00:01.0, 0x0000 -> 0000-00-01-0-0x0000
	// Replace colons and the dot in PCI address with hyphens.
	rfc1123PCIaddress := strings.ReplaceAll(strings.ReplaceAll(pciAddress, ":", "-"), ".", "-")
	newUID := fmt.Sprintf("%v-%v", rfc1123PCIaddress, pciid)

	return newUID
}

// ExtractCXINumber extracts the integer after '/dev/cxi' from a device path
func ExtractCXINumber(devicePath string) (int, error) {
	re := regexp.MustCompile(`/dev/cxi(\d+)`)
	matches := re.FindStringSubmatch(devicePath)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no CXI number found in device path: %s", devicePath)
	}
	return strconv.Atoi(matches[1])
}
