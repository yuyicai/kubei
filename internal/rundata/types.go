package rundata

import (
	"fmt"
	"sync"

	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"

	"github.com/yuyicai/kubei/pkg/ssh"
)

type Cluster struct {
	*Kubei
	Kubeadm *Kubeadm
	Mutex   sync.Mutex
}

func (c *Cluster) String() string {
	return fmt.Sprintf("%+v\n%+v", c.Kubei, c.Kubeadm)
}

type Kubei struct {
	ContainerEngine  ContainerEngine
	Kubernetes       Kubernetes
	ClusterNodes     ClusterNodes
	NetworkPlugins   NetworkPlugins
	HA               HA
	JumpServer       JumpServer
	Install          Install
	Reset            Reset
	Addons           Addons
	OfflineFile      string
	Online           bool
	CertNotAfterTime int
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

func NewCluster() *Cluster {
	return &Cluster{
		Kubei: &Kubei{
			ContainerEngine: ContainerEngine{},
			Kubernetes:      Kubernetes{},
			ClusterNodes:    ClusterNodes{},
			JumpServer:      JumpServer{},
			Install:         Install{},
			Reset:           Reset{},
			Addons:          Addons{},
		},
		Kubeadm: &Kubeadm{
			InitConfiguration: kubeadmapi.InitConfiguration{},
		},
	}
}
