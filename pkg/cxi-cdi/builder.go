// Package cxicdi provides Container Device Interface (CDI) specification building and management
// for HPE CXI devices. It creates and maintains CDI specs that describe how CXI devices
// should be made available to containers, including device nodes, mounts, and environment variables.
//
// CDI is a standardized way to describe how devices should be integrated into containers,
// providing a vendor-agnostic interface for container runtimes to access hardware devices.
package cxicdi

import (
	"fmt"

	device "github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/hpecxi"
	"k8s.io/klog/v2"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	// containerDevPath is the standard path where CXI devices appear inside containers
	containerDevPath = "/dev/cxi"
)

// getCXISpecs retrieves existing CDI specifications for CXI devices from the CDI cache.
// It filters the vendor specs to return only those matching the CXI device kind.
func getCXISpecs(cdiCache *cdiapi.Cache) []*cdiapi.Spec {
	cxiSpecs := []*cdiapi.Spec{}
	for _, cdiSpec := range cdiCache.GetVendorSpecs(device.Vendor) {
		if cdiSpec.Kind == device.Kind {
			cxiSpecs = append(cxiSpecs, cdiSpec)
		}
	}
	return cxiSpecs
}

// SyncRegistry synchronizes the CDI registry with detected CXI devices, mounts, and environment variables.
// It creates new CDI specifications or updates existing ones to reflect the current system state.
// The function handles both initial registry creation and incremental updates.
//
// Parameters:
// - cdiCache: CDI cache for managing specifications
// - detectedDevices: Currently available CXI devices on the system
// - detectedMounts: Required mounts for CXI device operation
// - envVars: Environment variables needed for CXI device access
// - doCleanup: Whether to remove absent devices from the registry
func SyncRegistry(cdiCache *cdiapi.Cache, detectedDevices device.DevicesInfo, detectedMounts []specs.Mount, envVars []string, doCleanup bool) error {

	vendorSpecs := getCXISpecs(cdiCache)
	devicesToAdd := detectedDevices.Clone()

	// If no existing specs, create a new registry from scratch
	if len(vendorSpecs) == 0 {
		klog.V(5).Infof("No existing specs found for vendor %v, creating new", device.Vendor)
		if err := buildNewRegistry(cdiCache, devicesToAdd, detectedMounts, envVars); err != nil {
			klog.V(5).Infof("Failed adding card to cdi registry: %v", err)
			return err
		}
		return nil
	}

	// TODO:
	// Update existing registry devices with detectedDevices.
	// Remove absent registry devices.

	return nil
}

// buildNewRegistry creates a new CDI registry specification from scratch with detected devices.
// It builds a complete CDI spec including device nodes, mounts, XPMEM support, and environment variables.
// The resulting spec is validated for CDI version compliance and written to the cache.
func buildNewRegistry(cdiCache *cdiapi.Cache, devices device.DevicesInfo, mounts []specs.Mount, envVars []string) error {
	klog.V(5).Infof("Adding %v devices to new spec", len(devices))

	spec := &specs.Spec{
		Kind: device.Kind,
	}

	// Add all detected devices to the specification
	addDevicesToSpec(devices, spec)
	klog.V(5).Infof("spec devices length: %v", len(spec.Devices))

	// Add required mounts (libraries, etc.)
	addMountstoSpec(mounts, spec)
	klog.V(5).Infof("spec mounts length: %v", len(spec.ContainerEdits.Mounts))

	// Add XPMEM support for cross-memory operations
	addXpmemtoSpec(spec)
	klog.V(5).Infof("spec xpmem device node added")

	// Add environment variables for proper runtime configuration
	addEnvVarsToSpec(envVars, spec)
	klog.V(5).Infof("spec environment variables added: %v", len(spec.ContainerEdits.Env))

	// Determine minimum required CDI version for this spec
	cdiVersion, err := cdiapi.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("failed to get minimum required CDI spec version: %v", err)
	}
	klog.V(5).Infof("CDI version required for new spec: %v", cdiVersion)
	spec.Version = cdiVersion

	// Generate a unique name for the specification
	specname, err := cdiapi.GenerateNameForSpec(spec)
	if err != nil {
		return fmt.Errorf("failed to generate name for cdi device spec: %+v", err)
	}
	klog.V(5).Infof("new name for new CDI spec: %v", specname)

	// Write the completed specification to the CDI cache
	err = cdiCache.WriteSpec(spec, specname)
	if err != nil {
		return fmt.Errorf("failed to write CDI spec %v: %v", specname, err)
	}

	return nil
}

// addDevicesToSpec adds CXI device nodes to the CDI specification.
// It creates individual device entries for each CXI device and an "all" device that includes all devices.
// Each device is configured as a character device with appropriate host and container paths.
func addDevicesToSpec(devices device.DevicesInfo, spec *specs.Spec) {
	devPath := device.GetDevPath()
	deviceNodes := []*specs.DeviceNode{}

	// Create individual device entries for each CXI device
	for _, device := range devices {
		deviceNode := &specs.DeviceNode{
			Path:     containerDevPath + fmt.Sprintf("%d", device.DeviceId),
			HostPath: devPath + fmt.Sprintf("%d", device.DeviceId),
			Type:     "c", // Character device
		}
		newDevice := specs.Device{
			Name: fmt.Sprintf("%d", device.DeviceId),
			ContainerEdits: specs.ContainerEdits{
				DeviceNodes: []*specs.DeviceNode{deviceNode},
			},
		}
		spec.Devices = append(spec.Devices, newDevice)
		deviceNodes = append(deviceNodes, deviceNode)
	}

	// Create an "all" device that includes all individual devices
	allDevice := specs.Device{
		Name: "all",
		ContainerEdits: specs.ContainerEdits{
			DeviceNodes: deviceNodes,
		},
	}
	spec.Devices = append(spec.Devices, allDevice)
}

// addMountstoSpec adds required mount points to the CDI specification.
// These mounts typically include library paths and other files needed for CXI device operation.
// Each mount is configured with default options for secure bind mounting.
func addMountstoSpec(mounts []specs.Mount, spec *specs.Spec) {
	for _, mount := range mounts {
		mount := &specs.Mount{
			HostPath:      mount.HostPath,
			ContainerPath: mount.ContainerPath,
			Options:       device.DefaultOptions,
			Type:          mount.Type,
		}
		spec.ContainerEdits.Mounts = append(spec.ContainerEdits.Mounts, mount)
	}
}

// addXpmemtoSpec adds XPMEM device node support to the CDI specification.
// XPMEM enables efficient cross-memory attach operations required for high-performance computing
// applications that use CXI devices for inter-process communication.
func addXpmemtoSpec(spec *specs.Spec) {
	xpmemPath := device.GetXpmemDevPath()
	xpmemMount := &specs.DeviceNode{
		Path:     xpmemPath,
		HostPath: xpmemPath,
		Type:     "c",
	}
	spec.ContainerEdits.DeviceNodes = append(spec.ContainerEdits.DeviceNodes, xpmemMount)
}

// addEnvVarsToSpec adds environment variables to the CDI specification.
// These environment variables are typically needed for proper library path configuration
// and other runtime requirements for CXI device access within containers.
func addEnvVarsToSpec(envVars []string, spec *specs.Spec) {
	if len(envVars) == 0 {
		klog.V(5).Infof("No environment variables to add to CDI spec")
		return
	}
	spec.ContainerEdits.Env = envVars
}
