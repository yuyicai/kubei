package options

type Kubeadm struct {
	Version              string
	ControlPlaneEndpoint string
	ImageRepository      string
	Networking           Networking
}

type Kubei struct {
	Reset            Reset
	ClusterNodes     ClusterNodes
	ContainerEngine  ContainerEngine
	Kubernetes       Kubernetes
	JumpServer       map[string]string
	OfflineFile      string
	CertNotAfterTime int
	NetworkType      string
}

type Kubernetes struct {
	Version string
}

type PublicHostInfo struct {
	Key      string
	User     string
	Password string
	Port     string
}

type ClusterNodes struct {
	PublicHostInfo PublicHostInfo

	Masters []string
	Workers []string
}

type ContainerEngine struct {
	Version string
}

type JumpServerHostInfo struct {
	PublicHostInfo
	Host string
}

type Reset struct {
	RemoveContainerEngine bool
	RemoveKubeComponent   bool
}

type Networking struct {
	ServiceSubnet string
	PodSubnet     string
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
	}

}
