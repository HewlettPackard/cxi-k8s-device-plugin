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

func main() {

	versions := [...]string{
		"HPE Slingshot device plugin for Kubernetes",
		fmt.Sprintf("%s version %s", os.Args[0], version),
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
	flag.Parse()

	for _, v := range versions {
		klog.Infof("%s", v)
	}

	l := plugin.HPECXILister{
		ResUpdateChan: make(chan dpm.PluginNameList),
		Heartbeat:     make(chan bool),
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
