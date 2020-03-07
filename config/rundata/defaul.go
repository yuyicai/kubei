package rundata

import "github.com/yuyicai/kubei/config/constants"

func DefaulkubeadmConf(k *Kubeadm) {

}

func DefaulKubeiConf(k *Kubei) {
	DefaulAddonsConf(&k.Addons)
}

func DefaulAddonsConf(a *Addons) {
	DefaulNetworkPluginsConf(&a.NetworkPlugins)
}

func DefaulNetworkPluginsConf(n *NetworkPlugins) {
	if n.Type == "" {
		n.Type = constants.DefaulNetworkPlugin
	}

	DefaulFlannelConf(&n.Flannel)
}

func DefaulFlannelConf(f *Flannel) {
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
