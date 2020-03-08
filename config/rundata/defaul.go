package rundata

import "github.com/yuyicai/kubei/config/constants"

func DefaulkubeadmConf(k *Kubeadm) {
	if k.LocalAPIEndpoint.BindPort==constants.IsNotSet{
		k.LocalAPIEndpoint.BindPort=constants.DefaultAPIBindPort
	}
}

func DefaulKubeiConf(k *Kubei) {
	defaulAddonsConf(&k.Addons)
}

func defaulAddonsConf(a *Addons) {
	defaulNetworkPluginsConf(&a.NetworkPlugins)
}

func defaulHAConf(h *HA) {
	if h.Type == constants.IsNotSet {
		h.Type = constants.HATypeNone
	}

	defaulLocalSLBConf(&h.LocalSLB)
}

func defaulNetworkPluginsConf(n *NetworkPlugins) {
	if n.Type == "" {
		n.Type = constants.DefaulNetworkPlugin
	}

	defaulFlannelConf(&n.Flannel)
}

func defaulFlannelConf(f *Flannel) {
	if f.BackendType == "" {
		f.BackendType = constants.DefaultFlannelBackendType
	}

	if f.Image.ImageRepository == "" {
		f.Image.ImageRepository = constants.DefaultFlannelImageRepository
	}

	if f.Image.ImageName == "" {
		f.Image.ImageName = constants.DefaultFlannelImageName
	}

	if f.Image.ImageTag == "" {
		f.Image.ImageTag = constants.DefaultFlannelVersion
	}
}

func defaulLocalSLBConf(l *LocalSLB) {
	if l.Type == constants.IsNotSet {
		l.Type = constants.LocalSLBTypeNginx
	}

	defaulNginxConf(&l.Nginx)
}

func defaulNginxConf(n *Nginx) {
	if n.Port == "" {
		n.Port = constants.DefaultNginxPort
	}

	if n.Image.ImageRepository == "" {
		n.Image.ImageName = constants.DefaultNginxImageRepository
	}

	if n.Image.ImageName == "" {
		n.Image.ImageName = constants.DefaultNginxImageName
	}

	if n.Image.ImageTag == "" {
		n.Image.ImageTag = constants.DefaultNginxVersion
	}
}
