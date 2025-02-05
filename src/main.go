package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.hpe.com/caio-davi/cxi-k8s-device-plugin/src/plugin"

	"github.com/golang/glog"
	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
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
	// this is also needed to enable glog usage in dpm
	flag.Parse()

	for _, v := range versions {
		glog.Infof("%s", v)
	}

	l := plugin.HPECXILister{
		ResUpdateChan: make(chan dpm.PluginNameList),
		Heartbeat:     make(chan bool),
	}
	manager := dpm.NewManager(&l)

	if pulse > 0 {
		go func() {
			glog.Infof("Heart beating every %d seconds", pulse)
			for {
				time.Sleep(time.Second * time.Duration(pulse))
				l.Heartbeat <- true
			}
		}()
	}

	go func() {
		// Check if there are Cassini NICs installed
		// TODO: check how many and update channel.
		var path = "/sys/class/cxi"
		if _, err := os.Stat(path); err == nil {
			l.ResUpdateChan <- []string{"cxi"}
		}
	}()
	manager.Run()

}
