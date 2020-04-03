package rundata

import "fmt"

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
