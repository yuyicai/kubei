package image

import (
	"archive/tar"
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

const (
	manifestFileName     = "manifest.json"
	legacyConfigFileName = "json"
	legacyLayerFileName  = "tar"
)

type Operator struct {
	SavePath     string
	CachePath    string
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
	ManifestItem ManifestItem
	Config       Config
	LayersInfo   []LayerInfo
	ConfigDigest digest.Digest
	configBytes  []byte
}

type ManifestItem struct {
	Config   string
	RepoTags []string
	Layers   []string
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

func NewImageOperator(imageUrl, savePath, cachePath string) (*Operator, error) {
	o := &Operator{}
	if err := o.setClient(imageUrl, "", ""); err != nil {
		return nil, err
	}
	o.SavePath = savePath
	o.CachePath = cachePath
	return o, nil
}

func NewImageOperatorSecure(imageUrl, savePath, cachePath, user, password string) (*Operator, error) {
	o := &Operator{}
	if err := o.setClient(imageUrl, user, password); err != nil {
		return nil, err
	}
	o.SavePath = savePath
	o.CachePath = cachePath
	return nil, nil
}

func (o *Operator) SaveImage() error {
	if err := o.DownloadLayers(); err != nil {
		return err
	}

	return o.saveImage()
}

func (o *Operator) DownloadLayers() error {
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
				decor.Name(blobInfo.Digest.Encoded()[:12]),
			),
			mpb.AppendDecorators(
				decor.Percentage(decor.WC{}),
				decor.Name(" ] "),
			),
		)

		blobInfo := blobInfo
		g.Go(func(ctx context.Context) error {
			return downloadLayer(o.RegistryInfo.client, o.Image.Repository, blobInfo, bar)
		})

	}

	return g.Wait()
}

func (o *Operator) checkImageUrl(imageUrl string) error {
	if strings.Contains(imageUrl, "@") {
		return errors.New("unsupported digest")
	}

	if !strings.HasPrefix(imageUrl, "http://") && !strings.HasPrefix(imageUrl, "https://") {
		imageUrl = fmt.Sprintf("https://%s", imageUrl)
	}

	registryURL, err := url.Parse(imageUrl)
	if err != nil {
		return err
	}

	o.RegistryInfo.Registry = registryURL.Host
	o.RegistryInfo.Scheme = registryURL.Scheme

	s := strings.SplitN(registryURL.Path, ":", 2)
	if len(s) != 2 {
		return errors.New("can not find tag")
	}

	o.Image.Repository = strings.Trim(s[0], "/")
	names := strings.Split(o.Image.Repository, "/")
	o.Image.Name = names[len(names)-1]
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

	if err := os.MkdirAll(o.CachePath, 0755); err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(o.CachePath, fmt.Sprintf("%s.%s",
		o.Image.ConfigDigest.Encoded(), legacyConfigFileName)))
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	_, err = file.Write(o.Image.configBytes)
	if err != nil {
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
		o.Image.LayersInfo[i].SaveFilePath = filepath.Join(o.CachePath, fmt.Sprintf("%s.%s",
			o.Image.LayersInfo[i].Digest.Encoded(), "tar"))
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

func (o *Operator) saveImage() error {
	file, err := os.Create(filepath.Join(o.SavePath,
		fmt.Sprintf("%s_%s.%s", o.Image.Name, o.Image.Tag, legacyLayerFileName)))
	if err != nil {
		return errors.WithStack(err)
	}
	defer file.Close()

	tw := tar.NewWriter(file)
	defer tw.Close()

	for _, l := range o.Image.LayersInfo {
		if err := sendLayerToTar(l, tw); err != nil {
			return err
		}
	}

	if err := tw.WriteHeader(&tar.Header{
		Name: fmt.Sprintf("%s.%s", o.Image.ConfigDigest.Encoded(), legacyConfigFileName),
		Mode: 644,
		Size: int64(len(o.Image.configBytes)),
	}); err != nil {
		return errors.WithStack(err)
	}
	_, err = tw.Write(o.Image.configBytes)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func downloadLayer(reg *registry.Registry, repo string, blobInfo LayerInfo, bar *mpb.Bar) error {
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

	file, err := os.Create(blobInfo.SaveFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, proxyReader)
	if err != nil {
		return err
	}

	return nil
}

func sendLayerToTar(layer LayerInfo, tw *tar.Writer) error {
	f, err := os.Open(layer.SaveFilePath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer f.Close()

	//gr, err := gzip.NewReader(f)
	//if err != nil {
	//	return errors.WithStack(err)
	//}
	//defer gr.Close()

	fi, err := f.Stat()
	if err != nil {
		return errors.WithStack(err)
	}

	hdr, err := tar.FileInfoHeader(fi, "")
	if err != nil {
		return errors.WithStack(err)
	}
	hdr.Mode = 644
	hdr.Name = fmt.Sprintf("%s.%s", layer.DiffID.Encoded(), legacyLayerFileName)

	if err := tw.WriteHeader(hdr); err != nil {
		return errors.WithStack(err)
	}

	_, err = io.Copy(tw, f)
	if err != nil {
		return errors.WithStack(err)
	}
	return nil
}
