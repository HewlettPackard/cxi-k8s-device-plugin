package test

import (
	"log"
	"os"
	"testing"

	"gopkg.in/yaml.v2"
	"tags.cncf.io/container-device-interface/specs-go"
)

func readCDIYAMLfile(t *testing.T, fileName string) specs.Spec {
	data, err := os.ReadFile(fileName)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// Unmarshal into specs.Spec
	var spec specs.Spec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}
	return spec
}

func getDevicesPath(t *testing.T, spec specs.Spec) map[string]string {
	paths := make(map[string]string)
	if len(spec.Devices) == 0 {
		log.Fatalf("No devices in the CDI specs.")
	}
	for _, device := range spec.Devices {
		if device.ContainerEdits.DeviceNodes != nil {
			for _, node := range device.ContainerEdits.DeviceNodes {
				paths[device.Name] = node.HostPath
			}
		}
	}
	return paths
}
func TestCxiSymlinksExist(t *testing.T) {
	devices := getDevicesPath(t, readCDIYAMLfile(t, "../tmp/cdi/hpe.com-cxi.yaml"))
	for deviceName, devicePath := range devices {
		if deviceName != "all" {
			_, err := os.Lstat(devicePath)
			if err != nil {
				t.Errorf("Failed to stat %s: %v", devicePath, err)
				continue
			}
		}
	}
}
