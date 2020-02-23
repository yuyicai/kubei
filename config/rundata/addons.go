package rundata

type Addons struct {
	Network Network
}

type Network struct {
	// network plugins, calico, flannel, none
	Type    string
	Flannel Flannel
	Calico  Calico
}

type Flannel struct {
	Image       FlannelImage
	BackendType string
}

type FlannelImage struct {
	Image string
}

type Calico struct {
	Image   CalicoImage
	Version string
}

type CalicoImage struct {
	Cni               string
	Typha             string
	Node              string
	KubeControllers   string
	Pod2daemonFlexvol string
}
