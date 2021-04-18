package registry

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
)

type ImageOperator struct {
	RegistryInfo RegistryInfo
	Image        Image
}

type RegistryInfo struct {
	Registry   string
	Repository string
	Scheme     string
	client     *registry.Registry
}

type Image struct {
	Tag          string
	Config       Config
	LayersInfo   []LayerInfo
	ConfigDigest digest.Digest
	configBytes  []byte
}

type LayerInfo struct {
	Digest       digest.Digest
	VerifyDigest digest.Digest
	Size         int64
	CheckFile    bool
	SaveFilePath string
}

type Config struct {
	RootFS RootFS `json:"rootfs,omitempty"`
}

type RootFS struct {
	Type    string          `json:"type"`
	DiffIDs []digest.Digest `json:"diff_ids,omitempty"`
}

func NewImageOperator(imageUrl string) (*ImageOperator, error) {
	o := &ImageOperator{}
	if err := o.setClient(imageUrl, "", ""); err != nil {
		return nil, err
	}
	return nil, nil
}

func NewImageOperatorSecure(imageUrl, user, password string) (*ImageOperator, error) {
	o := &ImageOperator{}
	if err := o.setClient(imageUrl, user, password); err != nil {
		return nil, err
	}

	return nil, nil
}

func (o *ImageOperator) checkImageUrl(imageUrl string) error {
	if strings.Contains(imageUrl, "@") {
		return errors.New("unsupported digest")
	}

	if !strings.HasPrefix(imageUrl, "http://") && !strings.HasPrefix(imageUrl, "https://") {
		imageUrl = fmt.Sprintf("https://%s", imageUrl)
	}

	registryUri, err := url.Parse(imageUrl)
	if err != nil {
		return err
	}

	o.RegistryInfo.Registry = registryUri.Host
	o.RegistryInfo.Scheme = registryUri.Scheme

	if !strings.Contains(registryUri.Path, ":") {
		return errors.New("can not find tag")
	}

	s := strings.SplitN(registryUri.Path, ":", 2)
	if len(s) != 2 {
		return errors.New("can not find tag")
	}

	o.RegistryInfo.Repository = strings.TrimPrefix(s[0], "/")
	o.Image.Tag = s[1]
	return err
}

func (o *ImageOperator) setClient(imageUrl, user, password string) error {
	if err := o.checkImageUrl(imageUrl); err != nil {
		return err
	}

	reg, err := New(fmt.Sprintf("%s://%s", o.RegistryInfo.Scheme, o.RegistryInfo.Registry), user, password)
	if err != nil {
		return errors.Wrapf(err, "failed to create registry client whit registry url: %s",
			o.RegistryInfo.Registry)
	}
	o.RegistryInfo.client = reg
	return nil
}

func (o *ImageOperator) downloadConfig() error {
	configBlob, err := o.RegistryInfo.client.DownloadBlob(o.RegistryInfo.Repository, o.Image.ConfigDigest)
	if err != nil {
		return err
	}
	defer configBlob.Close()

	o.Image.configBytes, err = ioutil.ReadAll(configBlob)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(o.Image.configBytes, &o.Image.Config); err != nil {
		return err
	}
	return nil
}

func (o *ImageOperator) SaveLayers() error {
	for _, blobInfo := range o.Image.LayersInfo {
		if err := saveLayer(o.RegistryInfo.client, o.RegistryInfo.Repository, blobInfo); err != nil {
			return err
		}
	}
	return nil
}

func saveLayer(reg *registry.Registry, repo string, blobInfo LayerInfo) error {
	if err := os.MkdirAll(filepath.Dir(blobInfo.SaveFilePath), 0755); err != nil {
		return err
	}

	blob, err := reg.DownloadBlob(repo, blobInfo.Digest)
	if err != nil {
		errors.Wrapf(err, "failed to download blob: %s/%s", repo, blobInfo.Digest)
	}
	defer blob.Close()

	gr, err := gzip.NewReader(blob)
	if err != nil {
		return err
	}
	defer gr.Close()

	file, err := os.Create(blobInfo.SaveFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, gr)
	if err != nil {
		return err
	}

	return nil
}
