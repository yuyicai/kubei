package rundata

import (
	"github.com/yuyicai/kubei/pkg/ssh"
)

type Kubei struct {
	Addons          Addons
	Reset           Reset
	ClusterNodes    ClusterNodes
	ContainerEngine ContainerEngine
	JumpServer      JumpServer
	IsHA            bool
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

type ContainerEngine struct {
	Version string
}

type Reset struct {
	RemoveContainerEngine bool
	RemoveKubeComponent   bool
}

type Kubeadm struct {
	Version              string
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
		ClusterNodes:    ClusterNodes{},
		ContainerEngine: ContainerEngine{},
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
