package cxicdi

import (
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	"tags.cncf.io/container-device-interface/specs-go"
)

type DeviceInfo struct {
	Name     string
	Path     string
	HostPath string
	Type     string
}

type MountInfo struct {
	HostPath      string
	ContainerPath string
	Options       []string
	Type          string
}

func readCDIYAMLfile(fileName string) specs.Spec {
	data, err := os.ReadFile(fileName)
	if err != nil {
		klog.Errorf("Failed to read YAML file. %v", err)
	}

	// Unmarshal into specs.Spec
	var spec specs.Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		klog.Errorf("Failed to unmarshal YAML: %v", err)
	}
	return spec
}

func getDevicesInfo(spec specs.Spec) map[string]DeviceInfo {
	devices := make(map[string]DeviceInfo)
	if len(spec.Devices) == 0 {
		klog.Error("No devices in the CDI specs.")
	}
	for _, device := range spec.Devices {
		if device.ContainerEdits.DeviceNodes != nil {
			for _, node := range device.ContainerEdits.DeviceNodes {
				devices[device.Name] = DeviceInfo{
					Name:     device.Name,
					Path:     node.Path,
					HostPath: node.HostPath,
					Type:     node.Type,
				}
			}
		}
	}
	return devices
}

func getMountsInfo(spec specs.Spec) []MountInfo {
	var mounts []MountInfo

	if spec.ContainerEdits.Mounts != nil {
		for _, mount := range spec.ContainerEdits.Mounts {
			mounts = append(mounts, MountInfo{
				HostPath:      mount.HostPath,
				ContainerPath: mount.ContainerPath,
				Options:       mount.Options,
				Type:          mount.Type,
			})
		}
	}
	return mounts
}

func ReadCXICDIYAML(fileName string) (map[string]DeviceInfo, []MountInfo) {
	spec := readCDIYAMLfile(fileName)
	devices := getDevicesInfo(spec)
	mounts := getMountsInfo(spec)

	if len(devices) == 0 {
		klog.Error("No devices found in the CDI spec.")
	}
	if len(mounts) == 0 {
		klog.Error("No mounts found in the CDI spec.")
	}

	return devices, mounts
}
