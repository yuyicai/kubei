package rundata

import "fmt"

type Image struct {
	ImageRepository string
	ImageName       string
	ImageTag        string
}

func (i *Image) GetImage() string {
	if i.ImageRepository == "" {
		return fmt.Sprintf("%s:%s", i.ImageName, i.ImageTag)
	}
	return fmt.Sprintf("%s/%s:%s", i.ImageRepository, i.ImageName, i.ImageTag)
}

type HA struct {
	// LocalSLB、None
	// TODO ExternalSLB
	Type     string
	LocalSLB LocalSLB
}

type LocalSLB struct {
	// Default Nginx
	// TODO HAproxy
	Type  string
	Nginx Nginx
}

type Nginx struct {
	Port  string
	Image Image
}
