package download

import (
	"fmt"
	"k8s.io/klog"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"

	"github.com/yuyicai/kubei/pkg/registry"
)

const (
	Registry   = "registry.aliyuncs.com"
	Repository = "kubebin"
	ImageName  = "kube-files"
)

func KubeFiles(tag, destPath string) error {
	imageUrl := fmt.Sprintf("%s:%s", path.Join(Registry, Repository, ImageName), tag)

	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	if destPath == "" {
		destPath = filepath.Join(home, ".kubei", tag)
	}

	klog.Infof("downloading %s.tar.gz to %s", ImageName, destPath)

	return registry.DownloadFile(imageUrl, "", "", destPath)
}
