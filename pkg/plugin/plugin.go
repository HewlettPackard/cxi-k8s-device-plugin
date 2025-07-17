package plugin

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"tags.cncf.io/container-device-interface/specs-go"

	cxicdi "github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/cxi-cdi"
	"github.com/HewlettPackard/cxi-k8s-device-plugin/pkg/hpecxi"

	"github.com/kubevirt/device-plugin-manager/pkg/dpm"
	"golang.org/x/net/context"
	"k8s.io/klog/v2"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const resourceNamespace string = "beta.hpe.com"

var envVars = map[string]string{
	"LD_LIBRARY_PATH": "/opt/cray/lib64:/usr/lib64",
}

// Plugin is identical to DevicePluginServer interface of device plugin API.
type HPECXIPlugin struct {
	CXIs       map[string]int
	Heartbeat  chan bool
	signal     chan os.Signal
	CDIEnabled bool
	CDIPath    string
	CDI        *specs.Spec
}

// Lister serves as an interface between imlementation and Manager machinery. User passes
// implementation of this interface to NewManager function. Manager will use it to obtain resource
// namespace, monitor available resources and instantate a new plugin for them.
type HPECXILister struct {
	ResUpdateChan chan dpm.PluginNameList
	Heartbeat     chan bool
	Signal        chan os.Signal
	CDIEnabled    bool
	CDIPath       string
}

func (l *HPECXILister) NewPlugin(resourceLastName string) dpm.PluginInterface {
	return &HPECXIPlugin{
		Heartbeat:  l.Heartbeat,
		CDIPath:    l.CDIPath,
		CDIEnabled: l.CDIEnabled,
	}
}

// Start is an optional interface that could be implemented by plugin.
// If case Start is implemented, it will be executed by Manager after
// plugin instantiation and before its registration to kubelet. This
// method could be used to prepare resources before they are offered
// to Kubernetes.
func (plugin *HPECXIPlugin) Start() error {
	plugin.signal = make(chan os.Signal, 1)
	if plugin.CDIEnabled {
		var err error
		plugin.CDI, err = cxicdi.GetCDISpecs(plugin.CDIPath)
		if err != nil {
			return err
		}
	}
	signal.Notify(plugin.signal, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	return nil
}

// Stop is an optional interface that could be implemented by plugin.
// If case Stop is implemented, it will be executed by Manager after the
// plugin is unregistered from kubelet. This method could be used to tear
// down resources.
func (p *HPECXIPlugin) Stop() error {
	return nil
}

func cxiSimpleHealthCheck(device *pluginapi.Device) string {
	var cxi *os.File
	var err error
	if cxi, err = os.Open("/dev/cxi" + device.ID); err != nil {
		klog.Error("Error opening /dev/cxi" + device.ID)
		return pluginapi.Unhealthy
	}
	cxi.Close()
	return pluginapi.Healthy
}

// GetDevicePluginOptions returns options to be communicated with Device
// Manager
func (p *HPECXIPlugin) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// PreStartContainer is expected to be called before each container start if indicated by plugin during registration phase.
// PreStartContainer allows kubelet to pass reinitialized devices to containers.
// PreStartContainer allows Device Plugin to run device specific operations on the Devices requested
func (plugin *HPECXIPlugin) PreStartContainer(ctx context.Context, r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}

// ListAndWatch returns a stream of List of Devices
// Whenever a Device state change or a Device disappears, ListAndWatch
// returns the new list
func (plugin *HPECXIPlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	if plugin.CXIs == nil {
		plugin.CXIs = make(map[string]int)
	}
	var devicesList = hpecxi.DiscoverDevices()
	for _, device := range devicesList {
		klog.Infof("Discovered device:  %s", device.Name)
		plugin.CXIs[device.Name] = int(device.DeviceId)
	}
	klog.Infof("Found %d HPE Slingshot NICs", len(plugin.CXIs))
	devs := make([]*pluginapi.Device, len(plugin.CXIs))
	func() {
		for _, id := range plugin.CXIs {
			dev := &pluginapi.Device{
				ID:     strconv.Itoa(id),
				Health: pluginapi.Healthy,
			}
			devs[id] = dev
		}
	}()

	s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})

loop:
	for {
		select {
		case <-plugin.Heartbeat:
			for i := 0; i < len(plugin.CXIs); i++ {
				devs[i].Health = cxiSimpleHealthCheck(devs[i])
				klog.Infof("[Health Check] cxi%d: %s", i, devs[i].Health)
			}
			s.Send(&pluginapi.ListAndWatchResponse{Devices: devs})
		case <-plugin.signal:
			klog.Infof("Received signal, exiting")
			break loop
		}
	}
	// returning a value with this function will unregister the plugin from k8s
	return nil
}

// GetPreferredAllocation returns a preferred set of devices to allocate
// from a list of available ones. The resulting preferred allocation is not
// guaranteed to be the allocation ultimately performed by the
// devicemanager. It is only designed to help the devicemanager make a more
// informed allocation decision when possible.
func (plugin *HPECXIPlugin) GetPreferredAllocation(context.Context, *pluginapi.PreferredAllocationRequest) (*pluginapi.PreferredAllocationResponse, error) {
	return &pluginapi.PreferredAllocationResponse{}, nil
}

// TODO:
// updateResponseForCDI updates the ContainerAllocateResponse with CDI specs
func (plugin *HPECXIPlugin) updateContainerAllocateResponseForCDI(car *pluginapi.ContainerAllocateResponse) {
	if !plugin.CDIEnabled {
		return
	}
	devices := cxicdi.GetDeviceSpecs(plugin.CDI)
	mounts := cxicdi.GetMounts(plugin.CDI)
	envVars := cxicdi.GetEnvVars(plugin.CDI)

	car.Devices = append(car.Devices, devices...)
	car.Mounts = append(car.Mounts, mounts...)
	car.Envs = envVars
}

// //
// Allocate is called during container creation so that the Device
// Plugin can run device specific operations and instruct Kubelet
// of the steps to make the Device available in the container
func (plugin *HPECXIPlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {
	var response pluginapi.AllocateResponse
	car := pluginapi.ContainerAllocateResponse{}

	if plugin.CDIEnabled {
		plugin.updateContainerAllocateResponseForCDI(&car)
	} else {
		var mountsList = hpecxi.DiscoverMounts()
		var devicesList = hpecxi.DiscoverDevices()

		car.Mounts = append(car.Mounts, cxicdi.ConvertMountstoMounts(mountsList)...)
		car.Devices = append(car.Devices, devicesList.ConvertToDeviceSpecs()...)
		car.Envs = envVars
	}

	response.ContainerResponses = append(response.ContainerResponses, &car)

	return &response, nil
}

// GetResourceNamespace must return namespace (vendor ID) of implemented Lister. e.g. for
// resources in format "color.example.com/<color>" that would be "color.example.com".
func (l *HPECXILister) GetResourceNamespace() string {
	return resourceNamespace
}

// Discover notifies manager with a list of currently available resources in its namespace.
// e.g. if "color.example.com/red" and "color.example.com/blue" are available in the system,
// it would pass PluginNameList{"red", "blue"} to given channel. In case list of
// resources is static, it would use the channel only once and then return. In case the list is
// dynamic, it could block and pass a new list each times resources changed. If blocking is
// used, it should check whether the channel is closed, i.e. Discover should stop.
func (l *HPECXILister) Discover(pluginListCh chan dpm.PluginNameList) {
	for {
		select {
		case newResourcesList := <-l.ResUpdateChan: // New resources found
			pluginListCh <- newResourcesList
		case <-pluginListCh: // Stop message received
			// Stop resourceUpdateCh
			return
		}
	}
}
