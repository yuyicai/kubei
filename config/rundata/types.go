package rundata

import (
	"github.com/yuyicai/kubei/pkg/ssh"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
)

type Configuration struct {
	Kubei   *Kubei
	Kubeadm *Kubeadm
}

type Kubei struct {
	ContainerEngine ContainerEngine
	Kubernetes      Kubernetes
	ClusterNodes    ClusterNodes
	NetworkPlugins  NetworkPlugins
	HA              HA
	JumpServer      JumpServer
	Install         Install
	Reset           Reset
	Addons          Addons
	OfflineFile     string
}

type JumpServer struct {
	*ssh.Client
	HostInfo HostInfo
}

type Reset struct {
	RemoveContainerEngine bool
	RemoveKubeComponent   bool
}

type Install struct {
	Type string
}

type Kubeadm struct {
	kubeadmapi.InitConfiguration
}

func NewKubei() *Kubei {
	return &Kubei{
		ContainerEngine: ContainerEngine{},
		Kubernetes:      Kubernetes{},
		ClusterNodes:    ClusterNodes{},
		JumpServer:      JumpServer{},
		Install:         Install{},
		Reset:           Reset{},
		Addons:          Addons{},
	}
}

func NewKubeadm() *Kubeadm {
	return &Kubeadm{
		InitConfiguration: kubeadmapi.InitConfiguration{},
	}
}
