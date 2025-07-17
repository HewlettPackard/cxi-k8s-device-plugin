package hpecxi

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"tags.cncf.io/container-device-interface/specs-go"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
)

type Env struct {
	Var []string `yaml:"env"`
}

var DefaultOptions = []string{"ro", "nosuid", "nodev", "bind", "relatime"}

func DiscoverDevices() map[string]*DeviceInfo {
	sysfsDir := path.Join(GetSysfsRoot(Sysfspath), Sysfspath)
	devices := make(map[string]*DeviceInfo)
	files, err := os.ReadDir(sysfsDir)
	if err != nil {
		if err == os.ErrNotExist {
			klog.V(1).Infof("No HPE CXI devices found on this host. %v does not exist", sysfsDir)
			return devices
		}
		klog.Errorf("could not read sysfs directory: %v", err)
		return devices
	}

	for _, card := range files {
		vendorID, err := os.ReadFile(path.Join(sysfsDir, card.Name(), "device/vendor"))
		if err != nil {
			klog.Errorf("Error reading device directory %s: %v", path.Join(sysfsDir, card.Name(), "device/vendor"), err)
		}

		if strings.TrimSpace(string(vendorID)) != HPEvendorID {
			continue
		}
		klog.V(4).Infof("Found HPE CXI PCI device: %s", card.Name())

		deviceDir := path.Join(sysfsDir, card.Name(), "device")
		deviceInfo := BuildDevice(deviceDir, card.Name())
		devices[deviceInfo.UID] = deviceInfo
	}

	return devices
}

func BuildDevice(deviceDir, deviceName string) *DeviceInfo {
	deviceInfo := &DeviceInfo{}
	devicePCIAddress := getPCISlot(path.Join(deviceDir, "uevent"))
	devicePCIID, err := os.ReadFile(path.Join(deviceDir, "device"))
	if err != nil {
		klog.Errorf("Failed to read PCI device ID for %s: %v", deviceName, err)
	}

	uid := DeviceUIDFromPCIinfo(devicePCIAddress, string(devicePCIID))

	deviceId, err := readDeviceNumber(deviceDir)
	if err != nil {
		klog.Errorf("Failed to read device number for %s: %v", deviceDir, err)
		os.Exit(1)
	}

	deviceInfo.Name = deviceName
	deviceInfo.UID = uid
	deviceInfo.PCIAddress = devicePCIAddress
	deviceInfo.DeviceId = deviceId
	deviceInfo.LocalCPUs = readPCIFile(deviceDir, "local_cpus")
	deviceInfo.LocalCPUList = readPCIFile(deviceDir, "local_cpulist")
	deviceInfo.NumaNode = readPCIFile(deviceDir, "numa_node")
	deviceInfo.Version = readPCIFile(deviceDir, "properties/cassini_version")
	deviceInfo.Speed = readPCIFile(deviceDir, "properties/speed")

	return deviceInfo
}

func readPCIFile(deviceDir string, propertyFileName string) string {
	devicePropertyFilePath := filepath.Join(deviceDir, propertyFileName)
	data, err := os.ReadFile(devicePropertyFilePath)
	if err != nil {
		klog.Errorf("Failed reading device file (%s): %+v", devicePropertyFilePath, err)
	}
	return strings.TrimSpace(string(data))
}

func readDeviceNumber(deviceDir string) (uint64, error) {
	contents, err := os.ReadDir(filepath.Join(deviceDir, "cxi"))
	if err != nil {
		klog.Errorf("Failed reading device from path (%s): %+v", deviceDir, err)
	}
	for _, item := range contents {
		name := filepath.Base(item.Name())
		if name[0:3] == "cxi" {
			id, _ := strconv.Atoi(name[len(name)-1:])
			klog.V(1).Infof("Found device %s with ID %d\n", name, id)
			return uint64(id), nil
		}
	}
	return 0, fmt.Errorf("cxi device number not found for %s", deviceDir)
}

func DiscoverMounts() []specs.Mount {
	mounts := make([]specs.Mount, 0)
	libFabricPath, err := GetLibfabricRoot()
	if err != nil {
		klog.Errorf("Failed to find Libfabric (%s): %+v", libFabricPath, err)
		os.Exit(1)
	}
	klog.V(1).Infof("Using Libfabric root: %s", libFabricPath)
	libcxiPath, err := GetLibcxiRoot()
	if err != nil {
		klog.Errorf("Failed to find Libcxi (%s): %+v", libcxiPath, err)
		os.Exit(1)
	}
	klog.V(1).Infof("Using Libcxi root: %s", libcxiPath)
	BuildMounts(libFabricPath, "libfabric", &mounts)
	BuildMounts(libcxiPath, "libcxi", &mounts)
	return mounts
}

func BuildMounts(path, name string, mounts *[]specs.Mount) {
	files, err := os.ReadDir(path)
	if err != nil {
		klog.Errorf("Failed to read directory (%s): %+v", path, err)
	}
	matchedFiles := make(map[string]string)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), name) {
			matchedFiles[file.Name()] = filepath.Join(path, file.Name())
		}
	}
	for _, path := range matchedFiles {
		mountInfo := specs.Mount{}
		mountInfo.HostPath = path
		mountInfo.ContainerPath = path

		mountType, err := os.Lstat(path)
		if err != nil {
			klog.Errorf("Failed to stat mount point (%s): %+v", path, err)
		}
		mode := mountType.Mode()
		switch {
		case mode.IsRegular():
			mountInfo.Type = "-"
		case mode.IsDir():
			mountInfo.Type = "d"
		case (mode & os.ModeSymlink) != 0:
			mountInfo.Type = "l"
		case (mode & os.ModeDevice) != 0:
			if (mode & os.ModeCharDevice) != 0 {
				mountInfo.Type = "c"
			} else {
				mountInfo.Type = "b"
			}
		case (mode & os.ModeNamedPipe) != 0:
			mountInfo.Type = "p"
		case (mode & os.ModeSocket) != 0:
			mountInfo.Type = "s"
		default:
			klog.Warningf("Unknown file type for %s", path)
		}
		*mounts = append(*mounts, mountInfo)
	}
}

func getPCISlot(filePath string) string {

	file, err := os.Open(filePath)
	if err != nil {
		klog.Errorf("Error opening file: %s", filePath)
	}
	defer file.Close()

	var pciSlotName string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "PCI_SLOT_NAME=") {
			pciSlotName = strings.TrimPrefix(line, "PCI_SLOT_NAME=")
		}
	}

	return pciSlotName
}

func DiscoverEnvVars(filePath string) []string {
	if filePath == "" {
		klog.V(1).Infof("No environment variables file specified, skipping discovery.")
		return nil
	}
	data, err := os.ReadFile(filePath)
	if err != nil {
		klog.Errorf("Failed to read `env-var` YAML file.")
	}

	var env Env
	if err := yaml.Unmarshal(data, &env); err != nil {
		klog.Errorf("Failed to unmarshal `env-var` YAML file.")
	}

	return env.Var
}
