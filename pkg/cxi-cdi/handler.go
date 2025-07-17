package cxicdi

import (
	"os"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"tags.cncf.io/container-device-interface/specs-go"
)

func GetCDISpecs(fileName string) (*specs.Spec, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		klog.Errorf("Failed to read CDI Spec file. %v", err)
		return nil, err
	}

	// Unmarshal into specs.Spec
	var spec specs.Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		klog.Errorf("Failed to unmarshal CDI Spec YAML: %v", err)
		return nil, err
	}
	return &spec, nil
}

func convertDeviceNodeToDeviceSpec(node specs.DeviceNode) *pluginapi.DeviceSpec {
	return &pluginapi.DeviceSpec{
		HostPath:      node.Path,
		ContainerPath: node.Path,
		Permissions:   "rw",
	}
}

func convertMountToMount(mount specs.Mount) *pluginapi.Mount {
	return &pluginapi.Mount{
		HostPath:      mount.HostPath,
		ContainerPath: mount.ContainerPath,
		ReadOnly:      true,
	}
}

func getCDIMounts(spec *specs.Spec) []specs.Mount {
	var mounts []specs.Mount

	if spec.ContainerEdits.Mounts != nil {
		for _, mount := range spec.ContainerEdits.Mounts {
			mounts = append(mounts, specs.Mount{
				HostPath:      mount.HostPath,
				ContainerPath: mount.ContainerPath,
				Options:       mount.Options,
				Type:          mount.Type,
			})
		}
	}
	return mounts
}

func getCDIDevices(spec *specs.Spec) map[string]specs.DeviceNode {
	devices := make(map[string]specs.DeviceNode)
	if len(spec.Devices) == 0 {
		klog.Error("No devices in the CDI specs.")
	}
	for _, device := range spec.Devices {
		if device.ContainerEdits.DeviceNodes != nil {
			for _, node := range device.ContainerEdits.DeviceNodes {
				devices[device.Name] = specs.DeviceNode{
					Path:     node.Path,
					HostPath: node.HostPath,
					Type:     node.Type,
				}
			}
		}
	}
	for _, deviceNode := range spec.ContainerEdits.DeviceNodes {
		if deviceNode != nil {
			devices[deviceNode.Path] = specs.DeviceNode{
				Path:     deviceNode.Path,
				HostPath: deviceNode.HostPath,
				Type:     deviceNode.Type,
			}
		}
	}

	return devices
}

func ConvertDeviceNodestoDeviceSpecs(deviceNodes map[string]specs.DeviceNode) []*pluginapi.DeviceSpec {
	var deviceSpecs []*pluginapi.DeviceSpec

	for _, node := range deviceNodes {
		deviceSpecs = append(deviceSpecs, convertDeviceNodeToDeviceSpec(node))
	}

	return deviceSpecs
}

func ConvertMountstoMounts(mounts []specs.Mount) []*pluginapi.Mount {
	var pluginMounts []*pluginapi.Mount

	for _, mount := range mounts {
		pluginMounts = append(pluginMounts, convertMountToMount(mount))
	}

	return pluginMounts
}

func GetDeviceSpecs(spec *specs.Spec) []*pluginapi.DeviceSpec {
	deviceNodes := getCDIDevices(spec)
	var deviceSpecs []*pluginapi.DeviceSpec

	for _, node := range deviceNodes {
		deviceSpecs = append(deviceSpecs, convertDeviceNodeToDeviceSpec(node))
	}

	return deviceSpecs
}

func GetMounts(spec *specs.Spec) []*pluginapi.Mount {
	mounts := getCDIMounts(spec)
	var pluginMounts []*pluginapi.Mount

	for _, mount := range mounts {
		pluginMounts = append(pluginMounts, convertMountToMount(mount))
	}

	return pluginMounts
}

func GetEnvVars(spec *specs.Spec) map[string]string {
	envVars := make(map[string]string)
	if spec.ContainerEdits.Env != nil {
		for _, env := range spec.ContainerEdits.Env {
			parts := strings.Split(env, "=")
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}
	return envVars
}
