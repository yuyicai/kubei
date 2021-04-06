package registry

import (
	"archive/tar"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"
	"k8s.io/klog"

	"github.com/yuyicai/kubei/pkg/util"
)

func DownloadImage(imageUrl, user, password, destPath string) error {
	img, err := checkImageUrl(imageUrl)
	if err != nil {
		return errors.Wrapf(err, "failed to check image url: %s", imageUrl)
	}

	hub, err := New(fmt.Sprintf("%s://%s", img.Scheme, img.Registry), user, password)
	if err != nil {
		return errors.Wrapf(err, "failed to create registry client whit registry url: %s", img.Registry)
	}
	return downloadImageFromRepository(hub, img.Repository, img.Tag, destPath)
}

func downloadImageFromRepository(hub *registry.Registry, repository, tag, destPath string) error {
	file := fmt.Sprintf("%s_%s", strings.ReplaceAll(repository, "/", "-"), tag)
	fw, err := os.Create(filepath.Join(destPath, file))
	if err != nil {
		return err
	}

	tw := tar.NewWriter(fw)
	defer tw.Close()

	_, err = hub.ManifestV2(repository, tag)
	if err != nil {
		return errors.Wrapf(err, "failed to get repository %s manifestV2", repository)
	}

	// todo download image layer
	return nil
}

func DownloadFile(imageUrl, user, password, destPath string) error {
	img, err := checkImageUrl(imageUrl)
	if err != nil {
		return errors.Wrapf(err, "failed to check image url: %s", imageUrl)
	}

	hub, err := New(fmt.Sprintf("%s://%s", img.Scheme, img.Registry), user, password)
	if err != nil {
		return errors.Wrapf(err, "failed to create registry client whit registry url: %s", img.Registry)
	}
	return downloadFileFromRepository(hub, img.Repository, img.Tag, destPath)
}

func downloadFileFromRepository(hub *registry.Registry, repository, tag, destPath string) error {

	manifestV2, err := hub.ManifestV2(repository, tag)
	if err != nil {
		return errors.Wrapf(err, "failed to get repository %s manifestV2", repository)
	}

	for _, layer := range manifestV2.Layers {
		klog.V(7).Infof("downloading layer: %v", layer)
		p := mpb.New(
			mpb.WithWidth(60),
			mpb.WithRefreshRate(180*time.Millisecond),
		)
		bar := p.Add(
			layer.Size,
			mpb.NewBarFiller("[=>-|"),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{}),
				decor.Name(" ] "),
				decor.EwmaSpeed(decor.UnitKiB, "% .2f", 60),
			),
		)
		blob, err := hub.DownloadBlob(repository, layer.Digest)
		if err != nil {
			errors.Wrapf(err, "failed to download blob: %s/%s", repository, layer.Digest)
		}
		proxyReader := bar.ProxyReader(blob)

		if err := util.DecompressToFile(proxyReader, destPath); err != nil {
			return errors.Wrapf(err, "failed to download blob: %s/%s", repository, layer.Digest)
		}

		blob.Close()
		proxyReader.Close()
	}
	return nil
}
