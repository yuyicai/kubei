# kubei

`kubei` (Kubernetes installer) 是一个go开发的用来部署Kubernetes高可用集群的命令行工具  

`kubei`原理：通过ssh连接到集群服务器，进行容器引擎安装、kubernetes组件安装、主机初始化配置、高可用负载均衡器配置、调用kubeadm初始化集群master、调用kubeadm将主机节点加入集群

# 功能
 - 一键部署高可用kubernetes集群
 - 离线部署 / 在线部署
 - 自定证书过期时间
 - 可使用普通用户部署安装(sudo用户)
 - 可使用跳板机连接主机部署安装

# 版本支持

| 应用/系统  |           版本            |
| :--------: | :-----------------------: |
| Kubernetes |  1.16.X、1.17.X、1.18.X   |
|  容器引擎  | Docker: 18.09.X、19.XX.XX |
|  网络插件  |      flannel: 0.11.0      |
|    系统    | Ubuntu16.04+、CentOS7.4+  |

*etcd版为kubeadm默认对应版本*

![k8s-ha](./docs/images/kube-ha.svg)

# 快速开始

|   主机    | 集群角色 |      系统版本      |
| :-------: | :------: | :----------------: |
| 10.3.0.10 |  master  | Ubuntu 18.04 LTS   |
| 10.3.0.11 |  master  | Ubuntu 18.04 LTS   |
| 10.3.0.12 |  master  | Ubuntu 18.04 LTS   |
| 10.3.0.20 |  worker  | Ubuntu 18.04 LTS   |
| 10.3.0.21 |  worker  | Ubuntu 18.04 LTS   |

*默认使用root用户和22端口，如果需要使用普通用户和其它ssh端口，请查看[ssh用户参数说明](./docs/flags.md)*

*如果要用密码做ssh登录验证，请查看[ssh用户参数说明](./docs/flags.md)*

**1、下载离线包：**

https://github.com/yuyicai/kubernetes-offline/releases

**2、下载部署程序**

https://github.com/yuyicai/kubei/releases

**3、执行部署命令：**

```
./kubei init \
 -k $HOME/.ssh/k8s.key \
 -m 10.3.0.10,10.3.0.11,10.3.0.12 \
 -n 10.3.0.20,10.3.0.21 \
 -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
```



[![asciicast](https://asciinema.org/a/353199.svg)](https://asciinema.org/a/353199)



[更多安装示例](./docs/example.md)、[参数说明](./docs/flags.md)



感谢：

[cobra]( https://github.com/spf13/cobra ): 命令框架采用`cobra`

[kubeadm]( https://github.com/kubernetes/kubernetes/blob/master/cmd/kubeadm/app/cmd/phases/workflow/doc.go ): 子命令工作流采用`kubeadm workflow`模块  

[kubespray]( https://github.com/kubernetes-sigs/kubespray/blob/master/docs/ha-mode.md ): 高可用配置采用`kubespray`项目的配置  



TODO

- [ ] calico网络组件支持
- [ ] 增加节点功能
- [x] 离线部署
- [x] 自定义证书过期时间