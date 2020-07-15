package rundata

import (
	"crypto"
	"crypto/x509"
	"github.com/pkg/errors"
	"github.com/yuyicai/kubei/config/constants"
	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmpkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"time"

	pkiutil "github.com/yuyicai/kubei/pkg/util/pki"
)

type ConfigMutatorsFunc func(*kubeadmapi.InitConfiguration, *pkiutil.CertConfig) error

// Cert represents a certificate that will create to function properly.
type Cert struct {
	Name     string
	LongName string
	BaseName string
	CAName   string
	Cert     *x509.Certificate
	Key      crypto.Signer
	// Some attributes will depend on the InitConfiguration, only known at runtime.
	// These functions will be run in series, passed both the InitConfiguration and a cert Config.
	ConfigMutators []ConfigMutatorsFunc
	Config         pkiutil.CertConfig
}

// GetConfig returns the definition for the given cert given the provided InitConfiguration
func (c *Cert) GetConfig(ic *kubeadmapi.InitConfiguration) (*pkiutil.CertConfig, error) {
	for _, f := range c.ConfigMutators {
		if err := f(ic, &c.Config); err != nil {
			return nil, err
		}
	}

	c.Config.PublicKeyAlgorithm = ic.ClusterConfiguration.PublicKeyAlgorithm()
	return &c.Config, nil
}

// CreateFromCA makes and writes a certificate using the given CA cert and key.
func (c *Cert) CreateFromCA(ic *kubeadmapi.InitConfiguration, caCert *x509.Certificate, caKey crypto.Signer) error {
	cfg, err := c.GetConfig(ic)
	if err != nil {
		return errors.Wrapf(err, "couldn't get configuration for %q certificate", c.Name)
	}
	c.Cert, c.Key, err = pkiutil.NewCertAndKey(caCert, caKey, cfg)
	if err != nil {
		return errors.Wrapf(err, "couldn't generate %q certificate", c.Name)
	}
	return nil
}

// CreateAsCA creates a certificate authority, writing the files to disk and also returning the created CA so it can be used to sign child certs.
func (c *Cert) CreateAsCA(ic *kubeadmapi.InitConfiguration) error {
	cfg, err := c.GetConfig(ic)
	if err != nil {
		return errors.Wrapf(err, "couldn't get configuration for %q CA certificate", c.Name)
	}
	c.Cert, c.Key, err = pkiutil.NewCertificateAuthority(cfg)
	if err != nil {
		return errors.Wrapf(err, "couldn't generate %q CA certificate", c.Name)
	}

	return nil
}

// CertificateTree is represents a one-level-deep tree, mapping a CA to the certs that depend on it.
type CertificateTree map[*Cert]Certificates

// CreateTree creates the CAs, certs signed by the CAs.
func (t CertificateTree) CreateTree(ic *kubeadmapi.InitConfiguration, notAfterTime time.Duration) error {
	for ca, leaves := range t {
		if ca.Cert == nil {

			if notAfterTime < constants.DefaultCertNotAfterTime {
				ca.Config.NotAfterTime = constants.DefaultCertNotAfterTime
			} else {
				ca.Config.NotAfterTime = notAfterTime
			}

			if err := ca.CreateAsCA(ic); err != nil {
				return err
			}
		}

		for _, leaf := range leaves {
			leaf.Config.NotAfterTime = notAfterTime
			if err := leaf.CreateFromCA(ic, ca.Cert, ca.Key); err != nil {
				return err
			}
		}
	}
	return nil
}

// CertificateMap is a flat map of certificates, keyed by Name.
type CertificateMap map[string]*Cert

// CertTree returns a one-level-deep tree, mapping a CA cert to an array of certificates that should be signed by it.
func (m CertificateMap) CertTree() (CertificateTree, error) {
	caMap := make(CertificateTree)

	for _, cert := range m {
		if cert.CAName == "" {
			if _, ok := caMap[cert]; !ok {
				caMap[cert] = []*Cert{}
			}
		} else {
			ca, ok := m[cert.CAName]
			if !ok {
				return nil, errors.Errorf("certificate %q references unknown CA %q", cert.Name, cert.CAName)
			}
			caMap[ca] = append(caMap[ca], cert)
		}
	}

	return caMap, nil
}

// AsMap returns the list of certificates as a map, keyed by name.
func (c Certificates) AsMap() CertificateMap {
	certMap := make(map[string]*Cert)
	for _, cert := range c {
		certMap[cert.Name] = cert
	}

	return certMap
}

var (
	// CertRootCA is the definition of the Kubernetes Root CA for the API Server and kubelet.
	CertRootCA = Cert{
		Name:     "ca",
		LongName: "self-signed Kubernetes CA to provision identities for other Kubernetes components",
		BaseName: kubeadmconstants.CACertAndKeyBaseName,
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: "kubernetes",
			},
		},
	}
	// CertAPIServer is the definition of the cert used to serve the Kubernetes API.
	CertAPIServer = Cert{
		Name:     "apiserver",
		LongName: "certificate for serving the Kubernetes API",
		BaseName: kubeadmconstants.APIServerCertAndKeyBaseName,
		CAName:   "ca",
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: kubeadmconstants.APIServerCertCommonName,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
			},
		},
		ConfigMutators: []ConfigMutatorsFunc{
			makeAltNamesMutator(kubeadmpkiutil.GetAPIServerAltNames),
		},
	}
	// CertKubeletClient is the definition of the cert used by the API server to access the kubelet.
	CertKubeletClient = Cert{
		Name:     "apiserver-kubelet-client",
		LongName: "certificate for the API server to connect to kubelet",
		BaseName: kubeadmconstants.APIServerKubeletClientCertAndKeyBaseName,
		CAName:   "ca",
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName:   kubeadmconstants.APIServerKubeletClientCertCommonName,
				Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}

	// CertFrontProxyCA is the definition of the CA used for the front end proxy.
	CertFrontProxyCA = Cert{
		Name:     "front-proxy-ca",
		LongName: "self-signed CA to provision identities for front proxy",
		BaseName: kubeadmconstants.FrontProxyCACertAndKeyBaseName,
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: "front-proxy-ca",
			},
		},
	}

	// CertFrontProxyClient is the definition of the cert used by the API server to access the front proxy.
	CertFrontProxyClient = Cert{
		Name:     "front-proxy-client",
		BaseName: kubeadmconstants.FrontProxyClientCertAndKeyBaseName,
		LongName: "certificate for the front proxy client",
		CAName:   "front-proxy-ca",
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: kubeadmconstants.FrontProxyClientCertCommonName,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}

	// CertEtcdCA is the definition of the root CA used by the hosted etcd server.
	CertEtcdCA = Cert{
		Name:     "etcd-ca",
		LongName: "self-signed CA to provision identities for etcd",
		BaseName: kubeadmconstants.EtcdCACertAndKeyBaseName,
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: "etcd-ca",
			},
		},
	}
	// CertEtcdServer is the definition of the cert used to serve etcd to clients.
	CertEtcdServer = Cert{
		Name:     "etcd-server",
		LongName: "certificate for serving etcd",
		BaseName: kubeadmconstants.EtcdServerCertAndKeyBaseName,
		CAName:   "etcd-ca",
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			},
		},
		ConfigMutators: []ConfigMutatorsFunc{
			makeAltNamesMutator(kubeadmpkiutil.GetEtcdAltNames),
			setCommonNameToNodeName(),
		},
	}
	// CertEtcdPeer is the definition of the cert used by etcd peers to access each other.
	CertEtcdPeer = Cert{
		Name:     "etcd-peer",
		LongName: "certificate for etcd nodes to communicate with each other",
		BaseName: kubeadmconstants.EtcdPeerCertAndKeyBaseName,
		CAName:   "etcd-ca",
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				Usages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			},
		},
		ConfigMutators: []ConfigMutatorsFunc{
			makeAltNamesMutator(kubeadmpkiutil.GetEtcdPeerAltNames),
			setCommonNameToNodeName(),
		},
	}
	// CertEtcdHealthcheck is the definition of the cert used by Kubernetes to check the health of the etcd server.
	CertEtcdHealthcheck = Cert{
		Name:     "etcd-healthcheck-client",
		LongName: "certificate for liveness probes to healthcheck etcd",
		BaseName: kubeadmconstants.EtcdHealthcheckClientCertAndKeyBaseName,
		CAName:   "etcd-ca",
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName:   kubeadmconstants.EtcdHealthcheckClientCertCommonName,
				Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}
	// CertEtcdAPIClient is the definition of the cert used by the API server to access etcd.
	CertEtcdAPIClient = Cert{
		Name:     "apiserver-etcd-client",
		LongName: "certificate the apiserver uses to access etcd",
		BaseName: kubeadmconstants.APIServerEtcdClientCertAndKeyBaseName,
		CAName:   "etcd-ca",
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName:   kubeadmconstants.APIServerEtcdClientCertCommonName,
				Organization: []string{kubeadmconstants.SystemPrivilegedGroup},
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}
)

// Certificates is a list of Certificates
type Certificates []*Cert

// GetDefaultCertList
func GetDefaultCertList() Certificates {
	return Certificates{
		&CertRootCA,
		&CertAPIServer,
		&CertKubeletClient,
		// Front Proxy certs
		&CertFrontProxyCA,
		&CertFrontProxyClient,
		// etcd certs
		&CertEtcdCA,
		&CertEtcdServer,
		&CertEtcdPeer,
		&CertEtcdHealthcheck,
		&CertEtcdAPIClient,
	}
}

func makeAltNamesMutator(f func(*kubeadmapi.InitConfiguration) (*certutil.AltNames, error)) ConfigMutatorsFunc {
	return func(mc *kubeadmapi.InitConfiguration, cc *pkiutil.CertConfig) error {
		altNames, err := f(mc)
		if err != nil {
			return err
		}
		cc.AltNames = *altNames
		return nil
	}
}

func setCommonNameToNodeName() ConfigMutatorsFunc {
	return func(mc *kubeadmapi.InitConfiguration, cc *pkiutil.CertConfig) error {
		cc.CommonName = mc.NodeRegistration.Name
		return nil
	}
}
