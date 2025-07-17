package cxicdi

import (
	"fmt"

	device "github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/hpecxi"
	"k8s.io/klog/v2"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"
	"tags.cncf.io/container-device-interface/specs-go"
)

const (
	containerDevPath = "/dev/cxi"
)

func getCXISpecs(cdiCache *cdiapi.Cache) []*cdiapi.Spec {
	cxiSpecs := []*cdiapi.Spec{}
	for _, cdiSpec := range cdiCache.GetVendorSpecs(device.Vendor) {
		if cdiSpec.Kind == device.Kind {
			cxiSpecs = append(cxiSpecs, cdiSpec)
		}
	}
	return cxiSpecs
}

func SyncRegistry(cdiCache *cdiapi.Cache, detectedDevices device.DevicesInfo, detectedMounts []specs.Mount, envVars []string, doCleanup bool) error {

	vendorSpecs := getCXISpecs(cdiCache)
	devicesToAdd := detectedDevices.Clone()

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

func buildNewRegistry(cdiCache *cdiapi.Cache, devices device.DevicesInfo, mounts []specs.Mount, envVars []string) error {
	klog.V(5).Infof("Adding %v devices to new spec", len(devices))

	spec := &specs.Spec{
		Kind: device.Kind,
	}

	addDevicesToSpec(devices, spec)
	klog.V(5).Infof("spec devices length: %v", len(spec.Devices))

	addMountstoSpec(mounts, spec)
	klog.V(5).Infof("spec mounts length: %v", len(spec.ContainerEdits.Mounts))

	addXpmemtoSpec(spec)
	klog.V(5).Infof("spec xpmem device node added")

	addEnvVarsToSpec(envVars, spec)
	klog.V(5).Infof("spec environment variables added: %v", len(spec.ContainerEdits.Env))

	cdiVersion, err := cdiapi.MinimumRequiredVersion(spec)
	if err != nil {
		return fmt.Errorf("failed to get minimum required CDI spec version: %v", err)
	}
	klog.V(5).Infof("CDI version required for new spec: %v", cdiVersion)
	spec.Version = cdiVersion

	specname, err := cdiapi.GenerateNameForSpec(spec)
	if err != nil {
		return fmt.Errorf("failed to generate name for cdi device spec: %+v", err)
	}
	klog.V(5).Infof("new name for new CDI spec: %v", specname)

	err = cdiCache.WriteSpec(spec, specname)
	if err != nil {
		return fmt.Errorf("failed to write CDI spec %v: %v", specname, err)
	}

	return nil
}

func addDevicesToSpec(devices device.DevicesInfo, spec *specs.Spec) {
	devPath := device.GetDevPath()
	deviceNodes := []*specs.DeviceNode{}
	for _, device := range devices {
		deviceNode := &specs.DeviceNode{
			Path:     containerDevPath + fmt.Sprintf("%d", device.DeviceId),
			HostPath: devPath + fmt.Sprintf("%d", device.DeviceId),
			Type:     "c",
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
	allDevice := specs.Device{
		Name: "all",
		ContainerEdits: specs.ContainerEdits{
			DeviceNodes: deviceNodes,
		},
	}
	spec.Devices = append(spec.Devices, allDevice)
}

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

func addXpmemtoSpec(spec *specs.Spec) {
	xpmemPath := device.GetXpmemDevPath()
	xpmemMount := &specs.DeviceNode{
		Path:     xpmemPath,
		HostPath: xpmemPath,
		Type:     "c",
	}
	spec.ContainerEdits.DeviceNodes = append(spec.ContainerEdits.DeviceNodes, xpmemMount)
}

func addEnvVarsToSpec(envVars []string, spec *specs.Spec) {
	if len(envVars) == 0 {
		klog.V(5).Infof("No environment variables to add to CDI spec")
		return
	}

	spec.ContainerEdits.Env = envVars

}
