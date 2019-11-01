package main
import (
	"net"//必须，51行创建sock
	"os"//必须， 116行os.Remove(m.socket)
	"path"//必须，82行path.Base(m.socket)
	"time"//创建连接时指定的间隔，一段时间没响应断开连接等等

	"golang.org/x/net/context"//必须，Allocate需要上下文信息
	"google.golang.org/grpc"//必须，grpc相关
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"//必须，proto文件
)
const(
	resourceName   = "demo.com/demo"//必须，部署时的资源名
	serverSock     = pluginapi.DevicePluginPath + "demo.sock"//必须，设备插件的sock
	//还需相关量用于Allocate
)
type DemoDevicePlugin struct {
	devs   []*pluginapi.Device
	socket string
	stop   chan interface{}
	health chan *pluginapi.Device
	server *grpc.Server
}
//GetDemoDevices区别：nvidia通过nvml获取，fpga读取相关设备文件
func NewDemoDevicePlugin() *DemoDevicePlugin {
	devs := GetDemoDevices()
	return &DemoDevicePlugin{
		devs:   devs,
		socket: serverSock,
		stop:   make(chan interface{}),
		health: make(chan *pluginapi.Device),
	}
}
//Serve完全一致，其中：Start、Register有所不同
func (m *DemoDevicePlugin) Serve() error {
	err := m.Start()
	if err != nil {
		return err
	}
	err = m.Register(pluginapi.KubeletSocket, resourceName)
	if err != nil {
		m.Stop()
		return err
	}
	return nil
}
func (m *DemoDevicePlugin) Start() error {
	err := m.cleanup()
	if err != nil {
		return err
	}
	sock, err := net.Listen("unix", m.socket)
	if err != nil {
		return err
	}
	m.server = grpc.NewServer([]grpc.ServerOption{}...)
	//需要列出所有方法，否则m报错
	pluginapi.RegisterDevicePluginServer(m.server, m)
	//以上实现完全一致，nvidia的协程部分做了进一步修饰
	go m.server.Serve(sock)
	//nvidia实现dial方法测试连接，fpga则通过阻塞sock连接测试服务是否活跃
	//考虑到Register方法中仍需要建立连接，因此需要dial，一致即可
	conn, err := dial(m.socket, 5*time.Second)
	if err != nil {
		return err
	}
	conn.Close()
	//nvidia此处进行健康检查
	// go m.healthcheck()
	return nil
}
//nvidia直接通过dial创建连接，fpga也类似dial实现创建连接
func (m *DemoDevicePlugin) Register(kubeletEndpoint, resourceName string) error {
	conn, err := dial(kubeletEndpoint, 5*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	//以下完全一致
	client := pluginapi.NewRegistrationClient(conn)
	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     path.Base(m.socket),
		ResourceName: resourceName,
	}
	_, err = client.Register(context.Background(), reqt)
	if err != nil {
		return err
	}
	return nil
}
//dial一致
func dial(unixSocketPath string, timeout time.Duration) (*grpc.ClientConn, error) {
	c, err := grpc.Dial(unixSocketPath, grpc.WithInsecure(), grpc.WithBlock(),
		grpc.WithTimeout(timeout),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}),
	)
	if err != nil {
		return nil, err
	}
	return c, nil
}
//Stop一致
func (m *DemoDevicePlugin) Stop() error{
	if m.server == nil {
		return nil
	}
	m.server.Stop()
	m.server = nil
	close(m.stop)
	return m.cleanup()
}
//cleanup一致
func (m *DemoDevicePlugin) cleanup() error {
	if err := os.Remove(m.socket); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
//以下实现ListAndWatch和Allocate
//发送设备信息，fpga此处进一步细化，实现功能基本一致
func (m *DemoDevicePlugin) ListAndWatch(e *pluginapi.Empty, s pluginapi.DevicePlugin_ListAndWatchServer) error {
	s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
	for {
		select {
		case <-m.stop:
			return nil
		case d := <-m.health:
			// FIXME: there is no way to recover from the Unhealthy state.
			d.Health = pluginapi.Unhealthy
			s.Send(&pluginapi.ListAndWatchResponse{Devices: m.devs})
		}
	}
}
//区别较大，存储设备信息采用的数据结构不同，以及proto文件定义略有不同
//总之根据proto定义选择合适存储结构
func (m *DemoDevicePlugin) Allocate(ctx context.Context, r *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse, error) {

	response := pluginapi.AllocateResponse{}
	return &response, nil
}
//未实现
func (m *DemoDevicePlugin) GetDevicePluginOptions(context.Context, *pluginapi.Empty) (*pluginapi.DevicePluginOptions, error) {
	return &pluginapi.DevicePluginOptions{}, nil
}
func (m *DemoDevicePlugin) PreStartContainer(context.Context, *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	return &pluginapi.PreStartContainerResponse{}, nil
}