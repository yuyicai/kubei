package options

import "github.com/yuyicai/kubei/config/rundata"

func (c *Kubeadm) ApplyTo(data *rundata.Kubeadm) {
	if c.ControlPlaneEndpoint != "" {
		data.ControlPlaneEndpoint = c.ControlPlaneEndpoint
	}

	if c.ImageRepository != "" {
		data.ImageRepository = c.ImageRepository
	}

	c.Networking.ApplyTo(&data.Networking)
}

func (c *Networking) ApplyTo(data *rundata.Networking) {
	if c.ServiceSubnet != "" {
		data.ServiceSubnet = c.ServiceSubnet
	}

	if c.PodSubnet != "" {
		data.PodSubnet = c.PodSubnet
	}
}
