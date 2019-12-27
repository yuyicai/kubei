package options

type Kubeadm struct {
	ControlPlaneEndpoint string
	ImageRepository      string
	Networking           Networking
}

type Kubei struct {
	ClusterNodes  ClusterNodes
	ContainerRuntime           ContainerRuntime
	KubeComponent KubeComponent
	JumpServer    map[string]string
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

type ContainerRuntime struct {
	Version string
}

type KubeComponent struct {
	Version string
}

type JumpServerHostInfo struct {
	PublicHostInfo
	Host string
}

type Networking struct {
	ServiceSubnet string
	PodSubnet     string
}

func NewKubei() *Kubei {
	return &Kubei{
		ClusterNodes:  ClusterNodes{},
		ContainerRuntime:           ContainerRuntime{},
		KubeComponent: KubeComponent{},
	}
}

func NewKubeadm() *Kubeadm {
	return &Kubeadm{
		ControlPlaneEndpoint: "",
		ImageRepository:      "",
		Networking:           Networking{},
	}

}
