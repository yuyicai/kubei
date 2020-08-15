package rundata

import "github.com/yuyicai/kubei/internal/constants"

func DefaultkubeadmCfg(k *Kubeadm, ki *Kubei) {
	if k.LocalAPIEndpoint.BindPort == 0 {
		k.LocalAPIEndpoint.BindPort = constants.DefaultAPIBindPort
	}

	setToEmptyString(&k.ClusterName, constants.DefaultClusterName)
	setToEmptyString(&k.Networking.DNSDomain, "cluster.local")

	if len(ki.ClusterNodes.Masters) > 0 {
		setToEmptyString(&k.LocalAPIEndpoint.AdvertiseAddress, ki.ClusterNodes.Masters[0].HostInfo.Host)
	}

}

func DefaultKubeiCfg(k *Kubei) {
	addonsCfg(&k.Addons)
	containerEngineCfg(&k.ContainerEngine)
	networkPluginsCfg(&k.NetworkPlugins)
	haCfg(&k.HA)
	clusterNodesCfg(&k.ClusterNodes)
	certCfg(&k.CertNotAfterTime)
}

func addonsCfg(a *Addons) {
}

func clusterNodesCfg(c *ClusterNodes) {
	for _, node := range c.GetAllNodes() {
		nodeCfg(node)
	}
}

func nodeCfg(node *Node) {
	if node.InstallType == "" {
		node.InstallType = constants.InstallTypeOnline
	}
}

func haCfg(h *HA) {
	if h.Type == "" {
		h.Type = constants.HATypeNone
	}

	localSLBCfg(&h.LocalSLB)
}

func networkPluginsCfg(n *NetworkPlugins) {
	if n.Type == "" {
		n.Type = constants.DefaulNetworkPlugin
	}

	flannelCfg(&n.Flannel)
}

func flannelCfg(f *Flannel) {
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

func localSLBCfg(l *LocalSLB) {
	if l.Type == "" {
		l.Type = constants.LocalSLBTypeNginx
	}

	nginxCfg(&l.Nginx)
}

func nginxCfg(n *Nginx) {
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

func containerEngineCfg(c *ContainerEngine) {
	if c.Type == "" {
		c.Type = constants.ContainerEngineTypeDocker
	}

	dockerCfg(&c.Docker)
}

func dockerCfg(d *Docker) {
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

func certCfg(t *int) {
	if *t == 0 {
		*t = constants.DefaultCertNotAfterYear
	}
}

func setToEmptyString(sp *string, s string) {
	if *sp == "" {
		*sp = s
	}
}
