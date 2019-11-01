package main
import(
	"fmt"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"//必须，proto文件
)
//nvidia通过nvml获取，fpga读取设备文件
func GetDemoDevices() []*pluginapi.Device {
	var devs = make([]*pluginapi.Device,1)
	for i, _ := range devs {
		devs[i] = &pluginapi.Device{
			ID:     fmt.Sprint(i),
			Health: pluginapi.Healthy,
		}
	}
	return devs
}