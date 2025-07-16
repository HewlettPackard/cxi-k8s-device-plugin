package cxicdi

import (
	"os"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"
	cdi "tags.cncf.io/container-device-interface/specs-go"
)

func GetCDISpecs(fileName string) cdi.Spec {
	data, err := os.ReadFile(fileName)
	if err != nil {
		klog.Errorf("Failed to read YAML file. %v", err)
	}

	// Unmarshal into specs.Spec
	var spec cdi.Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		klog.Errorf("Failed to unmarshal YAML: %v", err)
	}
	return spec
}

func GetDeviceNodes(spec cdi.Spec) map[string]cdi.DeviceNode {
	devices := make(map[string]cdi.DeviceNode)
	if len(spec.Devices) == 0 {
		klog.Error("No devices in the CDI specs.")
	}
	for _, device := range spec.Devices {
		if device.ContainerEdits.DeviceNodes != nil {
			for _, node := range device.ContainerEdits.DeviceNodes {
				devices[device.Name] = cdi.DeviceNode{
					Path:     node.Path,
					HostPath: node.HostPath,
					Type:     node.Type,
				}
			}
		}
	}
	return devices
}

func GetMounts(spec cdi.Spec) []cdi.Mount {
	var mounts []cdi.Mount

	if spec.ContainerEdits.Mounts != nil {
		for _, mount := range spec.ContainerEdits.Mounts {
			mounts = append(mounts, cdi.Mount{
				HostPath:      mount.HostPath,
				ContainerPath: mount.ContainerPath,
				Options:       mount.Options,
				Type:          mount.Type,
			})
		}
	}
	return mounts
}

// func ReadCXICDIYAML(fileName string) (map[string]cdi.DeviceNode, []cdi.Mount) {
// 	spec := GetCDISpecs(fileName)
// 	devices := getDevicesInfo(spec)
// 	mounts := getMountsInfo(spec)

// 	if len(devices) == 0 {
// 		klog.Error("No devices found in the CDI spec.")
// 	}
// 	if len(mounts) == 0 {
// 		klog.Error("No mounts found in the CDI spec.")
// 	}

// 	return devices, mounts
// }
