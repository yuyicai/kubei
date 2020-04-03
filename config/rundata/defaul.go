package rundata

import "github.com/yuyicai/kubei/config/constants"

func DefaultkubeadmConf(k *Kubeadm) {
	if k.LocalAPIEndpoint.BindPort == 0 {
		k.LocalAPIEndpoint.BindPort = constants.DefaultAPIBindPort
	}
}

func DefaultKubeiConf(k *Kubei) {
	defaultAddonsConf(&k.Addons)
	defaultContainerEngine(&k.ContainerEngine)
	defaultNetworkPluginsConf(&k.NetworkPlugins)
	defaultHAConf(&k.HA)
}

func defaultAddonsConf(a *Addons) {
}

func defaultHAConf(h *HA) {
	if h.Type == "" {
		h.Type = constants.HATypeNone
	}

	defaultLocalSLBConf(&h.LocalSLB)
}

func defaultNetworkPluginsConf(n *NetworkPlugins) {
	if n.Type == "" {
		n.Type = constants.DefaulNetworkPlugin
	}

	defaultFlannelConf(&n.Flannel)
}

func defaultFlannelConf(f *Flannel) {
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

func defaultLocalSLBConf(l *LocalSLB) {
	if l.Type == "" {
		l.Type = constants.LocalSLBTypeNginx
	}

	defaultNginxConf(&l.Nginx)
}

func defaultNginxConf(n *Nginx) {
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

func defaultContainerEngine(c *ContainerEngine) {
	if c.Type == "" {
		c.Type = constants.ContainerEngineTypeDocker
	}

	defaultDocker(&c.Docker)
}

func defaultDocker(d *Docker) {
	if d.CGroupDriver == "" {
		d.CGroupDriver = constants.DefaultCGroupDriver
	}

	if d.LogDriver == "" {
		d.LogDriver = constants.DefaultLogDriver
	}

	if d.LogOptsMaxSize == "" {
		d.LogOptsMaxSize = constants.DefaultLogOptsMaxSize
	}

	if d.StorageDriver == "" {
		d.StorageDriver = constants.DockerDefaultStorageDriver
	}
}
