package options

import (
	"github.com/yuyicai/kubei/internal/rundata"
)

func (c *Kubeadm) ApplyTo(data *rundata.Kubeadm) {
	if c.ControlPlaneEndpoint != "" {
		data.ControlPlaneEndpoint = c.ControlPlaneEndpoint
	}

	if c.ImageRepository != "" {
		data.ImageRepository = c.ImageRepository
	}

	c.Networking.ApplyTo(data)
}

func (c *Networking) ApplyTo(data *rundata.Kubeadm) {
	if c.ServiceSubnet != "" {
		data.Networking.ServiceSubnet = c.ServiceSubnet
	}

	if c.PodSubnet != "" {
		data.Networking.PodSubnet = c.PodSubnet
	}
}
