package registry

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/heroku/docker-registry-client/registry"
	"github.com/pkg/errors"
	"k8s.io/klog"
)

type image struct {
	Registry   string
	Repository string
	Tag        string
	Scheme     string
}

func New(registryURL, user, password string) (*registry.Registry, error) {
	if strings.Contains(registryURL, "https://") {
		return NewSecure(registryURL, user, password)
	}

	return NewInsecure(registryURL, user, password)
}

func NewSecure(registryURL, username, password string) (*registry.Registry, error) {
	transport := http.DefaultTransport
	return newFromTransport(registryURL, username, password, transport, registryLog)
}

func NewInsecure(registryURL, username, password string) (*registry.Registry, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	return newFromTransport(registryURL, username, password, transport, registryLog)
}

func registryLog(format string, args ...interface{}) {
	klog.V(8).Infof(format, args...)
}

func newFromTransport(registryURL, username, password string, transport http.RoundTripper, logf registry.LogfCallback) (*registry.Registry, error) {
	url := strings.TrimSuffix(registryURL, "/")
	transport = registry.WrapTransport(transport, url, username, password)
	r := &registry.Registry{
		URL: url,
		Client: &http.Client{
			Transport: transport,
		},
		Logf: logf,
	}

	if err := r.Ping(); err != nil {
		return nil, err
	}
	return r, nil
}

func CheckImageUrl(imageUrl string) (image, error) {
	img := image{}

	if strings.Contains(imageUrl, "@") {
		return img, errors.New("unsupported digest")
	}

	if !strings.HasPrefix(imageUrl, "http://") && !strings.HasPrefix(imageUrl, "https://") {
		imageUrl = fmt.Sprintf("https://%s", imageUrl)
	}

	registryUri, err := url.Parse(imageUrl)
	if err != nil {
		return img, err
	}

	img.Registry = registryUri.Host
	img.Scheme = registryUri.Scheme

	if !strings.Contains(registryUri.Path, ":") {
		return img, errors.New("can not find tag")
	}

	s := strings.SplitN(registryUri.Path, ":", 2)
	if len(s) != 2 {
		return img, errors.New("can not find tag")
	}

	img.Repository = strings.TrimPrefix(s[0], "/")
	img.Tag = s[1]
	return img, err
}
