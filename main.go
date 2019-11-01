package main
import (
	"log"
	"os"//必须，系统监听
	"syscall"//必须，系统调用

	"github.com/fsnotify/fsnotify"//必须，FS监听
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"//必须，proto文件
)
func main(){
	//newFSWatcher完全一致
	watcher, err := newFSWatcher(pluginapi.DevicePluginPath)
	if err != nil {
		os.Exit(1)
	}
	defer watcher.Close()
	//newOSWatcher完全一致
	sigs := newOSWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	restart := true
	var devicePlugin *DemoDevicePlugin//结构体有所不同
L:
	for {
		if restart {
			if devicePlugin != nil {
				devicePlugin.Stop()//Stop完全一致
			}
			//方法中对于设备的获取有所不同，本质都是给Device结构体赋值
			devicePlugin = NewDemoDevicePlugin()
			//Serve完全一致，其中的方法：Start、Register有所不同
			if err := devicePlugin.Serve(); err != nil {
				log.Printf("inotify: %s,retrying...", err)
			} else {
				restart = false
			}
		}
		//以下内容基本一致，可根据情况增加case，如fpga插件
		select {
		case event := <-watcher.Events:
			if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
				log.Printf("inotify: %s created, restarting.", pluginapi.KubeletSocket)
				restart = true
			}
		case err := <-watcher.Errors:
			log.Printf("inotify: %s", err)
		case s := <-sigs:
			switch s {
			case syscall.SIGHUP:
				log.Println("Received SIGHUP, restarting.")
				restart = true
			default:
				log.Printf("Received signal \"%v\", shutting down.", s)
				devicePlugin.Stop()
				break L
			}
		}
	}
}
