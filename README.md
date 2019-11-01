# k8s-demo-device-plugin
通过学习k8s设备插件相关内容，进行简单总结
## 说明
项目通过dep进行依赖管理
* 在`%GOPATH%/src`中创建项目文件夹project
* 通过goland打开project
* 在goland下方的终端执行`dep init`，生成vendor文件夹以及lock和toml
* 单独下载好需要的包，如`github.com/xxx`，解压到vendor中
* 在project中创建源文件,`import ("github.com/xxx")`即可
* 此时将光标定位到`xxx`，`ctrl+b`即可跳转至vendor下对应的文件(夹)
## 参考
* [k8s-fpga-device-plugin](https://github.com/Xilinx/FPGA_as_a_Service)
* [k8s-nvidia-device-plugin](https://github.com/NVIDIA/k8s-device-plugin)
* [k8s-dumb-device-plugin](https://github.com/everpeace/k8s-dumb-device-plugin)
