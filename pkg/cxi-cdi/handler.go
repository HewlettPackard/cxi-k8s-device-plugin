package cxicdi

import (
	"os"
	"strings"

	"gopkg.in/yaml.v2"
	"k8s.io/klog/v2"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"tags.cncf.io/container-device-interface/specs-go"
)

// GetCDISpecs reads and parses a CDI specification file from disk.
// It loads a YAML file containing CDI device specifications and unmarshals it
// into a specs.Spec structure for use by the device plugin.
//
// Parameters:
// - fileName: Path to the CDI specification YAML file
//
// Returns the parsed CDI specification or an error if reading/parsing fails.
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

// convertDeviceNodeToDeviceSpec converts a CDI DeviceNode to a Kubernetes DeviceSpec.
// This translation enables CDI device nodes to be used by the Kubernetes device plugin API.
// The device is configured with read-write permissions for container access.
func convertDeviceNodeToDeviceSpec(node specs.DeviceNode) *pluginapi.DeviceSpec {
	return &pluginapi.DeviceSpec{
		HostPath:      node.Path,
		ContainerPath: node.Path,
		Permissions:   "rw",
	}
}

// getCDIDevices extracts device nodes from a CDI specification.
// It processes both individual device entries and container-level device nodes,
// creating a map of device names to their corresponding device node specifications.
// This handles the CDI structure where devices can be defined at multiple levels.
func getCDIDevices(spec *specs.Spec) map[string]specs.DeviceNode {
	devices := make(map[string]specs.DeviceNode)
	if len(spec.Devices) == 0 {
		klog.Error("No devices in the CDI specs.")
	}

	// Process individual device entries and their device nodes
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

	// Process container-level device nodes (global device nodes)
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

// ConvertDeviceNodestoDeviceSpecs converts a map of CDI device nodes to Kubernetes DeviceSpecs.
// This function is used to transform CDI device specifications into the format required
// by the Kubernetes device plugin API for device allocation responses.
func ConvertDeviceNodestoDeviceSpecs(deviceNodes map[string]specs.DeviceNode) []*pluginapi.DeviceSpec {
	var deviceSpecs []*pluginapi.DeviceSpec

	for _, node := range deviceNodes {
		deviceSpecs = append(deviceSpecs, convertDeviceNodeToDeviceSpec(node))
	}

	return deviceSpecs
}

// GetDeviceSpecs extracts and converts device specifications from a CDI spec.
// It processes the CDI specification to identify all device nodes and converts them
// to the Kubernetes DeviceSpec format required for device allocation responses.
// This is the main interface for retrieving device specifications from CDI.
func GetDeviceSpecs(spec *specs.Spec) []*pluginapi.DeviceSpec {
	deviceNodes := getCDIDevices(spec)
	var deviceSpecs []*pluginapi.DeviceSpec

	for _, node := range deviceNodes {
		deviceSpecs = append(deviceSpecs, convertDeviceNodeToDeviceSpec(node))
	}

	return deviceSpecs
}

// convertCDIMountToSpecMount converts a CDI Mount specification to a Kubernetes Mount.
// This translation allows CDI mount specifications to be used by the Kubernetes device plugin API.
// Mounts are configured as read-only for security by default.
func convertCDIMountToSpecMount(mount specs.Mount) *pluginapi.Mount {
	return &pluginapi.Mount{
		HostPath:      mount.HostPath,
		ContainerPath: mount.ContainerPath,
		ReadOnly:      true,
	}
}

// ConvertCDIMountsToSpecMounts converts CDI mount specifications to Kubernetes mounts.
// This function transforms CDI mount specs into the format expected by the
// Kubernetes device plugin API for container allocation responses.
func ConvertCDIMountsToSpecMounts(mounts []specs.Mount) []*pluginapi.Mount {
	var pluginMounts []*pluginapi.Mount

	for _, mount := range mounts {
		pluginMounts = append(pluginMounts, convertCDIMountToSpecMount(mount))
	}

	return pluginMounts
}

// getCDIMounts extracts mount specifications from a CDI spec.
// It processes both container-level mounts and device-specific mounts,
// returning a consolidated list of all required mount points.
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

// GetMounts extracts and converts mount specifications from a CDI spec.
// It processes the CDI specification to identify all required mounts and converts them
// to the Kubernetes Mount format required for device allocation responses.
// This includes library paths and other files needed for device operation.
func GetMounts(spec *specs.Spec) []*pluginapi.Mount {
	mounts := getCDIMounts(spec)
	var pluginMounts []*pluginapi.Mount

	for _, mount := range mounts {
		pluginMounts = append(pluginMounts, convertCDIMountToSpecMount(mount))
	}

	return pluginMounts
}

// GetEnvVars extracts environment variables from a CDI specification.
// It parses environment variable strings in "KEY=VALUE" format from the CDI spec
// and returns them as a map suitable for use in Kubernetes device allocation responses.
// These environment variables are typically needed for proper library path configuration.
func GetEnvVars(spec *specs.Spec) map[string]string {
	envVars := make(map[string]string)
	if spec.ContainerEdits.Env != nil {
		// Parse environment variables in "KEY=VALUE" format
		for _, env := range spec.ContainerEdits.Env {
			parts := strings.Split(env, "=")
			if len(parts) == 2 {
				envVars[parts[0]] = parts[1]
			}
		}
	}
	return envVars
}
