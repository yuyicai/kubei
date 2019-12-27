package rundata

import (
	"github.com/yuyicai/kubei/pkg/ssh"
)

type Kubei struct {
	ClusterNodes     ClusterNodes
	ContainerRuntime ContainerRuntime
	Kube             KubeComponent
	JumpServer       JumpServer
	IsHA             bool
}

type ClusterNodes struct {
	Masters []*Node
	Worker  []*Node
}

const (
	Apt = 1 << iota
	Yum
	Offline
)

type Node struct {
	SSH              *ssh.Client
	HostInfo         HostInfo
	Name             string
	InstallationType int
}

type JumpServer struct {
	*ssh.Client
	IsUse    bool
	HostInfo HostInfo
}

type HostInfo struct {
	Host     string
	User     string
	Password string
	Port     string
	Key      string
}

type ContainerRuntime struct {
	Version string
}

type KubeComponent struct {
	Version string
}

type Kubeadm struct {
	ControlPlaneEndpoint string
	ImageRepository      string
	Networking           Networking
	Token                Token
}

type Networking struct {
	ServiceSubnet string
	PodSubnet     string
}

type Token struct {
	Token          string
	CaCertHash     string
	CertificateKey string
}

func NewKubei() *Kubei {
	return &Kubei{
		ClusterNodes:     ClusterNodes{},
		ContainerRuntime: ContainerRuntime{},
		Kube:             KubeComponent{},
	}
}

func NewKubeadm() *Kubeadm {
	return &Kubeadm{
		ControlPlaneEndpoint: "",
		ImageRepository:      "",
		Networking:           Networking{},
		Token:                Token{},
	}
}
