package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"k8s.io/klog/v2"
	cdiapi "tags.cncf.io/container-device-interface/pkg/cdi"

	cxicdi "github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/cxi-cdi"
	device "github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/hpecxi"
)

var (
	version      string
	verboseLevel int
)

func main() {

	klog.InitFlags(nil)
	defer klog.Flush()

	command := newCommand()
	err := command.Execute()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cxi-cdi-generator [--cdi-dir=<cdi directory>] [--dry-run]",
		Short: "HPE Slingshot CDI Generator",
		Long:  "HPE Slingshot CDI Generator detects supported NICs and creates CDI specs for them.",

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			flag.Set("V", strconv.Itoa(verboseLevel))
			_ = flag.CommandLine.Parse([]string{})
		},

		RunE: cobraRunFunc,
	}

	cmd.Version = version
	cmd.PersistentFlags().IntVarP(&verboseLevel, "verbose", "V", 5, "Set verbosity level for logging (0-5)")
	cmd.Flags().BoolP("version", "v", false, "Show the version of the binary")
	cmd.Flags().String("cdi-dir", "/etc/cdi", "CDI spec directory")
	cmd.Flags().String("env-vars", "", "YAML file with environment variables to set in CDI specs")
	cmd.Flags().BoolP("dry-run", "n", false, "Dry-run, do not create CDI manifests")
	cmd.SetVersionTemplate("HPE Slingshot CDI Generator Version: {{.Version}}\n")

	return cmd
}

func cobraRunFunc(cmd *cobra.Command, args []string) error {
	cdiDir := cmd.Flag("cdi-dir").Value.String()
	envVarsPath := cmd.Flag("env-vars").Value.String()

	klog.V(1).Infof("Refreshing CDI registry")
	if err := cdiapi.Configure(cdiapi.WithSpecDirs(cdiDir)); err != nil {
		fmt.Printf("unable to refresh the CDI registry: %v", err)
		return err
	}

	cdiCache, err := cdiapi.NewCache(cdiapi.WithAutoRefresh(false), cdiapi.WithSpecDirs(cdiDir))
	if err != nil {
		return err
	}

	dryRun := false
	if cmd.Flag("dry-run").Value.String() == "true" {
		dryRun = true
	}

	var devicesList = device.DiscoverDevices()
	for _, device := range devicesList {
		klog.V(1).Infof("Discovered device: %s, PCI Address: %s\n", device.UID, device.PCIAddress)
	}

	var mountsList = device.DiscoverMounts()
	for _, mount := range mountsList {
		klog.V(1).Infof("Discovered mount: %s, Path: %s\n", mount.Name, mount.HostPath)
	}

	var envVars = device.DiscoverEnvVars(envVarsPath)
	klog.V(1).Infof("Discovered environment variables from %s:\n", envVarsPath)
	for _, envVar := range envVars {
		klog.V(1).Infof("->" + envVar)
	}

	if dryRun {
		return nil
	}

	if err := cxicdi.SyncRegistry(cdiCache, devicesList, mountsList, envVars, true); err != nil {
		klog.Errorf("unable to sync detected devices to CDI registry: %v", err)
		return err
	}

	if err := cdiCache.Refresh(); err != nil {
		return err
	}

	// Fix CDI spec permissions as the default permission (600) prevents
	// use without root or sudo:
	// https://github.com/cncf-tags/container-device-interface/issues/224
	specs := cdiCache.GetVendorSpecs(device.Vendor)
	for _, spec := range specs {
		if err := os.Chmod(spec.GetPath(), 0o644); err != nil {
			return err
		}
	}

	return nil
}
