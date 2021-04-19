package image

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/go-kratos/kratos/pkg/sync/errgroup"
	"github.com/heroku/docker-registry-client/registry"
	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/vbauerster/mpb/v6"
	"github.com/vbauerster/mpb/v6/decor"

	pkgreg "github.com/yuyicai/kubei/pkg/registry"
)

type Operator struct {
	SavePath     string
	RegistryInfo RegistryInfo
	Image        Image
}

type RegistryInfo struct {
	Registry string
	Scheme   string
	User     string
	Password string
	client   *registry.Registry
}

type Image struct {
	Repository   string
	Name         string
	Tag          string
	Config       Config
	LayersInfo   []LayerInfo
	ConfigDigest digest.Digest
	configBytes  []byte
}

type LayerInfo struct {
	distribution.Descriptor
	DiffID       digest.Digest
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

func NewImageOperator(imageUrl, savePath string) (*Operator, error) {
	o := &Operator{}
	if err := o.setClient(imageUrl, "", ""); err != nil {
		return nil, err
	}
	o.SavePath = savePath
	return o, nil
}

func NewImageOperatorSecure(imageUrl, savePath, user, password string) (*Operator, error) {
	o := &Operator{}
	if err := o.setClient(imageUrl, user, password); err != nil {
		return nil, err
	}
	o.SavePath = savePath
	return nil, nil
}

func (o *Operator) checkImageUrl(imageUrl string) error {
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

	o.Image.Repository = strings.TrimPrefix(s[0], "/")
	o.Image.Tag = s[1]
	return err
}

func (o *Operator) setClient(imageUrl, user, password string) error {
	if err := o.checkImageUrl(imageUrl); err != nil {
		return err
	}

	reg, err := pkgreg.New(fmt.Sprintf("%s://%s", o.RegistryInfo.Scheme, o.RegistryInfo.Registry), user, password)
	if err != nil {
		return errors.Wrapf(err, "failed to create registry client whit registry url: %s",
			o.RegistryInfo.Registry)
	}
	o.RegistryInfo.client = reg
	return nil
}

func (o *Operator) downloadConfig() error {
	configBlob, err := o.RegistryInfo.client.DownloadBlob(o.Image.Repository, o.Image.ConfigDigest)
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

func (o *Operator) setLayersInfo(layers []distribution.Descriptor) {
	o.Image.LayersInfo = make([]LayerInfo, len(layers))
	for i, m := range layers {
		o.Image.LayersInfo[i].Descriptor = m
		o.Image.LayersInfo[i].CheckFile = true
		o.Image.LayersInfo[i].DiffID = o.Image.Config.RootFS.DiffIDs[i]
		o.Image.LayersInfo[i].SaveFilePath = filepath.Join(o.SavePath, fmt.Sprintf("%s.%s",
			o.Image.LayersInfo[i].DiffID.Encoded(), "tar"))
	}
}

func (o *Operator) setImageManifestInfo() error {
	manifestV2, err := o.RegistryInfo.client.ManifestV2(o.Image.Repository, o.Image.Tag)
	if err != nil {
		return errors.Wrapf(err, "failed to get repository %s manifestV2", o.Image.Repository)
	}

	o.Image.ConfigDigest = manifestV2.Config.Digest

	if err := o.downloadConfig(); err != nil {
		return err
	}

	if len(manifestV2.Layers) != len(o.Image.Config.RootFS.DiffIDs) {
		return errors.Errorf("bad image manifest info with %s", o.Image.Repository)
	}

	o.setLayersInfo(manifestV2.Layers)

	return nil
}

func (o *Operator) SaveLayers() error {
	if err := o.setImageManifestInfo(); err != nil {
		return err
	}

	p := mpb.New(
		mpb.WithWidth(80),
		mpb.WithRefreshRate(180*time.Millisecond),
	)

	g := errgroup.WithCancel(context.Background())
	g.GOMAXPROCS(5)

	for _, blobInfo := range o.Image.LayersInfo {
		bar := p.Add(
			blobInfo.Size,
			mpb.NewBarFiller("[=>-|"),
			mpb.PrependDecorators(
				decor.Name(fmt.Sprintf("digest:%s diffID:%s",
					blobInfo.Digest.Encoded()[:12], blobInfo.DiffID.Encoded()[:12])),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{}),
				decor.Name(" ] "),
			),
		)

		blobInfo := blobInfo
		g.Go(func(ctx context.Context) error {
			return saveLayer(o.RegistryInfo.client, o.Image.Repository, blobInfo, bar)
		})

	}

	return g.Wait()
}

func saveLayer(reg *registry.Registry, repo string, blobInfo LayerInfo, bar *mpb.Bar) error {
	if err := os.MkdirAll(filepath.Dir(blobInfo.SaveFilePath), 0755); err != nil {
		return err
	}

	blob, err := reg.DownloadBlob(repo, blobInfo.Digest)
	if err != nil {
		errors.Wrapf(err, "failed to download blob: %s/%s", repo, blobInfo.Digest)
	}
	defer blob.Close()

	proxyReader := bar.ProxyReader(blob)
	defer proxyReader.Close()

	gr, err := gzip.NewReader(proxyReader)
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
