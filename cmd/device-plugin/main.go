package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/hpecxi"
	"github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/plugin"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"k8s.io/klog/v2"
)

var version string

type cdiFlag struct {
	set   bool
	value string
}

func (c *cdiFlag) String() string {
	if !c.set {
		return ""
	}
	return c.value
}

func (c *cdiFlag) Set(val string) error {
	c.set = true
	if val == "" {
		c.value = "/etc/cdi/"
	} else {
		c.value = val
	}
	return nil
}

func main() {

	versions := [...]string{
		"HPE Slingshot device plugin for Kubernetes",
		fmt.Sprintf("%s version %s", os.Args[0], version),
	}

	for i, arg := range os.Args {
		if arg == "-enable-cdi" || arg == "--enable-cdi" {
			os.Args[i] = "-enable-cdi=/etc/cdi/"
		}
	}

	flag.Usage = func() {
		for _, v := range versions {
			fmt.Fprintf(os.Stderr, "%s\n", v)
		}
		fmt.Fprintln(os.Stderr, "Usage:")
		flag.PrintDefaults()
	}
	var pulse int
	flag.IntVar(&pulse, "pulse", 0, "time between health check polling in seconds.  Set to 0 to disable.")
	var cdi cdiFlag
	flag.Var(&cdi, "enable-cdi", "enable CDI and set CDI path (default: /etc/cdi/ when flag is present)")
	flag.Parse()

	if cdi.set {
		klog.Infof("CDI is enabled with path: %s\n", cdi.value)
	} else {
		klog.Info("CDI is not enabled. Using discovery only.")
	}

	for _, v := range versions {
		klog.Infof("%s", v)
	}

	l := plugin.HPECXILister{
		ResUpdateChan: make(chan dpm.PluginNameList),
		Heartbeat:     make(chan bool),
		CDI:           cdi.value,
	}
	manager := dpm.NewManager(&l)

	if pulse > 0 {
		go func() {
			klog.Infof("Heart beating every %d seconds", pulse)
			for {
				time.Sleep(time.Second * time.Duration(pulse))
				l.Heartbeat <- true
			}
		}()
	}

	go func() {
		// Check if there are Cassini NICs installed
		// TODO: check how many and update channel.
		var path = path.Join(hpecxi.GetSysfsRoot(hpecxi.Sysfspath), hpecxi.Sysfspath)
		if _, err := os.Stat(path); err == nil {
			l.ResUpdateChan <- []string{"cxi"}
		}
	}()
	manager.Run()

}
