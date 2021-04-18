package download

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/yuyicai/kubei/pkg/image"
	"path"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const (
	Registry        = "registry.aliyuncs.com"
	Repository      = "kubebin"
	ImageName       = "kube-files"
	KubeFileSubPath = ".kubei"
)

func KubeFiles(tag, destPath string) error {
	imageUrl := fmt.Sprintf("%s:%s", path.Join(Registry, Repository, ImageName), tag)

	home, err := homedir.Dir()
	if err != nil {
		return err
	}
	if destPath == "" {
		destPath = filepath.Join(home, KubeFileSubPath, tag)
	}

	color.HiBlack("Downloading %s.tar.gz to %s", ImageName, destPath)
	if err := image.DownloadFile(imageUrl, "", "", destPath); err != nil {
		return err
	}
	color.HiGreen("done✅️")
	return nil
}
