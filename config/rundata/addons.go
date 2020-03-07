package rundata

import "fmt"

type Addons struct {
	NetworkPlugins NetworkPlugins
	HA             HA
}

type NetworkPlugins struct {
	// network plugins, calico, flannel, none
	Type    string
	Flannel Flannel
	Calico  Calico
}

type Flannel struct {
	Image       Image
	BackendType string
}

type Calico struct {
	Image Image
}

func (c *Calico) GetImage(image string) string {
	if c.Image.ImageRepository == "" {
		return fmt.Sprintf("%s:%s", c.Image.ImageName, c.Image.ImageTag)
	}
	return fmt.Sprintf("%s/%s:%s", c.Image.ImageRepository, image, c.Image.ImageTag)
}

type HA struct {
	// LocalSLB、None
	// TODO ExternalSLB
	Type     int
	LocalSLB LocalSLB
}

type LocalSLB struct {
	// Default Nginx
	// TODO HAproxy
	Type  int
	Nginx Nginx
}

type Nginx struct {
	Image Image
	Port  string
}
