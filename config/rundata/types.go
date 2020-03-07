package rundata

import (
	"fmt"
	"github.com/yuyicai/kubei/pkg/ssh"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
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

type Image struct {
	ImageRepository string
	ImageName       string
	ImageTag        string
}

func (i *Image) GetImage() string {
	if i.ImageRepository == "" {
		return fmt.Sprintf("%s:%s", i.ImageName, i.ImageTag)
	}
	return fmt.Sprintf("%s/%s:%s", i.ImageRepository, i.ImageName, i.ImageTag)
}

type Kubeadm struct {
	kubeadmapi.InitConfiguration
	Token   Token
	Version string
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
		InitConfiguration: kubeadmapi.InitConfiguration{},
		Token:             Token{},
		Version:           "",
	}
}
