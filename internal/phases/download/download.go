package download

import "github.com/yuyicai/kubei/pkg/registry"

func Download(imageUrl,user,password,destPath string) error  {
	return registry.DownloadFile(imageUrl,user,password,destPath)
}
