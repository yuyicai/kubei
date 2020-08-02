package rundata

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"k8s.io/apimachinery/pkg/util/validation"
	"net"
	"time"

	"github.com/pkg/errors"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	certutil "k8s.io/client-go/util/cert"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/features"
	kubeadmutil "k8s.io/kubernetes/cmd/kubeadm/app/util"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"

	"github.com/yuyicai/kubei/config/constants"
	pkiutil "github.com/yuyicai/kubei/pkg/pki"
)

type ConfigMutatorsFunc func(*Node, *kubeadmapi.InitConfiguration, *pkiutil.CertConfig) error

// Cert represents a certificate that will create to function properly.
type Cert struct {
	Name         string
	LongName     string
	BaseName     string
	CAName       string
	IsKubeConfig bool
	Cert         *x509.Certificate
	Key          crypto.Signer
	// Some attributes will depend on the InitConfiguration, only known at runtime.
	// These functions will be run in series, passed both the InitConfiguration and a cert Config.
	ConfigMutators []ConfigMutatorsFunc
	Config         pkiutil.CertConfig
	KubeConfig     clientcmdapi.Config
}

// GetConfig returns the definition for the given cert given the provided InitConfiguration
func (c *Cert) GetConfig(node *Node, ic *kubeadmapi.InitConfiguration) (*pkiutil.CertConfig, error) {
	for _, f := range c.ConfigMutators {
		if err := f(node, ic, &c.Config); err != nil {
			return nil, err
		}
	}

	c.Config.PublicKeyAlgorithm = ic.ClusterConfiguration.PublicKeyAlgorithm()
	return &c.Config, nil
}

// CreateFromCA makes and writes a certificate using the given CA cert and key.
func (c *Cert) CreateFromCA(node *Node, ic *kubeadmapi.InitConfiguration, caCert *x509.Certificate, caKey crypto.Signer) error {
	cfg, err := c.GetConfig(node, ic)
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
func (c *Cert) CreateAsCA(node *Node, ic *kubeadmapi.InitConfiguration) error {
	cfg, err := c.GetConfig(node, ic)
	if err != nil {
		return errors.Wrapf(err, "couldn't get configuration for %q CA certificate", c.Name)
	}
	c.Cert, c.Key, err = pkiutil.NewCertificateAuthority(cfg)
	if err != nil {
		return errors.Wrapf(err, "couldn't generate %q CA certificate", c.Name)
	}

	return nil
}

func (c *Cert) CreateKubeConfig(ic *kubeadmapi.InitConfiguration, caCert *x509.Certificate) error {

	if !c.IsKubeConfig {
		return nil
	}

	encodedClientKey, err := pkiutil.EncodePrivateKeyPEM(c.Key)
	if err != nil {
		return err
	}

	controlPlaneEndpoint, err := kubeadmutil.GetControlPlaneEndpoint(ic.ControlPlaneEndpoint, &ic.LocalAPIEndpoint)
	if err != nil {
		return err
	}

	c.KubeConfig = *kubeconfigutil.CreateWithCerts(
		controlPlaneEndpoint,
		ic.ClusterName,
		c.Config.CommonName,
		pkiutil.EncodeCertPEM(caCert),
		encodedClientKey,
		pkiutil.EncodeCertPEM(c.Cert),
	)
	return nil
}

// CertificateTree is represents a one-level-deep tree, mapping a CA to the certs that depend on it.
type CertificateTree map[*Cert]Certificates

// Create creates the CAs, certs signed by the CAs.
func (t CertificateTree) Create(node *Node, ic *kubeadmapi.InitConfiguration, notAfterTime time.Duration) error {
	for ca, leaves := range t {
		if ca.Cert == nil {

			if notAfterTime < constants.DefaultCertNotAfterTime {
				ca.Config.NotAfterTime = constants.DefaultCertNotAfterTime
			} else {
				ca.Config.NotAfterTime = notAfterTime
			}

			if err := ca.CreateAsCA(node, ic); err != nil {
				return err
			}
		}

		for _, leaf := range leaves {
			leaf.Config.NotAfterTime = notAfterTime
			if err := leaf.CreateFromCA(node, ic, ca.Cert, ca.Key); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t CertificateTree) CreateKubeConfig(ic *kubeadmapi.InitConfiguration) error {
	for ca, certs := range t {
		if ca.Name == "ca" {
			for _, cert := range certs {
				if cert.IsKubeConfig {
					if err := cert.CreateKubeConfig(ic, ca.Cert); err != nil {
						return err
					}
				}
			}
			return nil
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
			makeAltNamesMutator(GetAPIServerAltNames),
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
			makeAltNamesMutator(GetEtcdAltNames),
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
			makeAltNamesMutator(GetEtcdPeerAltNames),
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

	// CertAPIServerKubeletClient is the definition of the cert used by the kubelet to access the API server.
	CertAPIServerKubeletClient = Cert{
		Name:         "kubelet",
		LongName:     "certificate for the kubelet to access the API server",
		BaseName:     kubeadmconstants.KubeletKubeConfigFileName,
		CAName:       "ca",
		IsKubeConfig: true,
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName:   kubeadmconstants.APIServerKubeletClientCertCommonName,
				Organization: []string{kubeadmconstants.NodesGroup},
				Usages:       []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
		ConfigMutators: []ConfigMutatorsFunc{
			setCommonNameToKubelet(),
		},
	}

	// CertAPIServerKubeletClient is the definition of the cert used by the kubelet to access the API server.
	CertAPIServerControllerManagerClient = Cert{
		Name:         "controller-manager",
		LongName:     "certificate for the kube-controller-manager to access the API server",
		BaseName:     kubeadmconstants.ControllerManagerKubeConfigFileName,
		CAName:       "ca",
		IsKubeConfig: true,
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: kubeadmconstants.ControllerManagerUser,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}

	// CertAPIServerKubeletClient is the definition of the cert used by the kubelet to access the API server.
	CertAPIServerSchedulerClient = Cert{
		Name:         "scheduler",
		LongName:     "certificate for the kube-scheduler to access the API server",
		BaseName:     kubeadmconstants.SchedulerKubeConfigFileName,
		CAName:       "ca",
		IsKubeConfig: true,
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName: kubeadmconstants.SchedulerUser,
				Usages:     []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
			},
		},
	}

	// CertAPIServerKubeletClient is the definition of the cert used by the kubelet to access the API server.
	CertAPIServerAdminClient = Cert{
		Name:         "admin",
		LongName:     "certificate for the kube-admin to access the API server",
		BaseName:     kubeadmconstants.AdminKubeConfigFileName,
		CAName:       "ca",
		IsKubeConfig: true,
		Config: pkiutil.CertConfig{
			Config: certutil.Config{
				CommonName:   "kubernetes-admin",
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
	certRootCA := CertRootCA
	certAPIServer := CertAPIServer
	certKubeletClient := CertKubeletClient
	certAPIServerControllerManagerClient := CertAPIServerControllerManagerClient
	certAPIServerSchedulerClient := CertAPIServerSchedulerClient
	certAPIServerAdminClient := CertAPIServerAdminClient
	//certAPIServerKubeletClient := CertAPIServerKubeletClient
	// Front Proxy certs
	certFrontProxyCA := CertFrontProxyCA
	certFrontProxyClient := CertFrontProxyClient
	// etcd certs
	certEtcdCA := CertEtcdCA
	certEtcdServer := CertEtcdServer
	certEtcdPeer := CertEtcdPeer
	certEtcdHealthcheck := CertEtcdHealthcheck
	certEtcdAPIClient := CertEtcdAPIClient

	return Certificates{
		&certRootCA,
		&certAPIServer,
		&certKubeletClient,
		&certAPIServerControllerManagerClient,
		&certAPIServerSchedulerClient,
		&certAPIServerAdminClient,
		//&certAPIServerKubeletClient,
		// Front Proxy certs
		&certFrontProxyCA,
		&certFrontProxyClient,
		// etcd certs
		&certEtcdCA,
		&certEtcdServer,
		&certEtcdPeer,
		&certEtcdHealthcheck,
		&certEtcdAPIClient,
	}
}

func makeAltNamesMutator(f func(*Node, *kubeadmapi.InitConfiguration) (*certutil.AltNames, error)) ConfigMutatorsFunc {
	return func(node *Node, mc *kubeadmapi.InitConfiguration, cc *pkiutil.CertConfig) error {
		altNames, err := f(node, mc)
		if err != nil {
			return err
		}
		cc.AltNames = *altNames
		return nil
	}
}

func setCommonNameToNodeName() ConfigMutatorsFunc {
	return func(node *Node, mc *kubeadmapi.InitConfiguration, cc *pkiutil.CertConfig) error {
		cc.CommonName = node.Name
		return nil
	}
}

func setCommonNameToKubelet() ConfigMutatorsFunc {
	return func(node *Node, mc *kubeadmapi.InitConfiguration, cc *pkiutil.CertConfig) error {
		cc.CommonName = fmt.Sprintf("%s%s", kubeadmconstants.NodesUserPrefix, node.Name)
		return nil
	}
}

// GetAPIServerAltNames builds an AltNames object for to be used when generating apiserver certificate
func GetAPIServerAltNames(node *Node, cfg *kubeadmapi.InitConfiguration) (*certutil.AltNames, error) {
	// host address
	host := net.ParseIP(node.HostInfo.Host)
	if host == nil {
		return nil, errors.Errorf("error parsing node host %v: is not a valid textual representation of an IP address",
			node.HostInfo.Host)
	}

	internalAPIServerVirtualIP, err := kubeadmconstants.GetAPIServerVirtualIP(cfg.Networking.ServiceSubnet, features.Enabled(cfg.FeatureGates, features.IPv6DualStack))
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get first IP address from the given CIDR: %v", cfg.Networking.ServiceSubnet)
	}

	// create AltNames with defaults DNSNames/IPs
	altNames := &certutil.AltNames{
		DNSNames: []string{
			"localhost",
			"kubernetes",
			"kubernetes.default",
			"kubernetes.default.svc",
			"kubernetes.default.svc.cluster.local",
			fmt.Sprintf("kubernetes.default.svc.%s", cfg.Networking.DNSDomain),
		},
		IPs: []net.IP{
			net.IPv4(127, 0, 0, 1),
			net.IPv6loopback,
			internalAPIServerVirtualIP,
			host,
		},
	}

	// add cluster controlPlaneEndpoint if present (dns or ip)
	if len(cfg.ControlPlaneEndpoint) > 0 {
		if host, _, err := kubeadmutil.ParseHostPort(cfg.ControlPlaneEndpoint); err == nil {
			if ip := net.ParseIP(host); ip != nil {
				altNames.IPs = append(altNames.IPs, ip)
			} else {
				altNames.DNSNames = append(altNames.DNSNames, host)
			}
		} else {
			return nil, errors.Wrapf(err, "error parsing cluster controlPlaneEndpoint %q", cfg.ControlPlaneEndpoint)
		}
	}

	exAltNames := append(cfg.APIServer.CertSANs, node.Name)
	appendSANsToAltNames(altNames, exAltNames, kubeadmconstants.APIServerCertName)

	return altNames, nil
}

// appendSANsToAltNames parses SANs from as list of strings and adds them to altNames for use on a specific cert
// altNames is passed in with a pointer, and the struct is modified
// valid IP address strings are parsed and added to altNames.IPs as net.IP's
// RFC-1123 compliant DNS strings are added to altNames.DNSNames as strings
// RFC-1123 compliant wildcard DNS strings are added to altNames.DNSNames as strings
// certNames is used to print user facing warnings and should be the name of the cert the altNames will be used for
func appendSANsToAltNames(altNames *certutil.AltNames, SANs []string, certName string) {
	for _, altname := range SANs {
		if ip := net.ParseIP(altname); ip != nil {
			altNames.IPs = append(altNames.IPs, ip)
		} else if len(validation.IsDNS1123Subdomain(altname)) == 0 {
			altNames.DNSNames = append(altNames.DNSNames, altname)
		} else if len(validation.IsWildcardDNS1123Subdomain(altname)) == 0 {
			altNames.DNSNames = append(altNames.DNSNames, altname)
		} else {
			fmt.Printf(
				"[certificates] WARNING: '%s' was not added to the '%s' SAN, because it is not a valid IP or RFC-1123 compliant DNS entry\n",
				altname,
				certName,
			)
		}
	}
}

// GetEtcdPeerAltNames builds an AltNames object for generating the etcd peer certificate.
// Hostname and `API.AdvertiseAddress` are included if the user chooses to promote the single node etcd cluster into a multi-node one (stacked etcd).
// The user can override the listen address with `Etcd.ExtraArgs` and add SANs with `Etcd.PeerCertSANs`.
func GetEtcdPeerAltNames(node *Node, cfg *kubeadmapi.InitConfiguration) (*certutil.AltNames, error) {
	return getAltNames(node, cfg, kubeadmconstants.EtcdPeerCertName)
}

// GetEtcdAltNames builds an AltNames object for generating the etcd server certificate.
// `advertise address` and localhost are included in the SAN since this is the interfaces the etcd static pod listens on.
// The user can override the listen address with `Etcd.ExtraArgs` and add SANs with `Etcd.ServerCertSANs`.
func GetEtcdAltNames(node *Node, cfg *kubeadmapi.InitConfiguration) (*certutil.AltNames, error) {
	return getAltNames(node, cfg, kubeadmconstants.EtcdServerCertName)
}

// getAltNames builds an AltNames object with the cfg and certName.
func getAltNames(node *Node, cfg *kubeadmapi.InitConfiguration, certName string) (*certutil.AltNames, error) {
	// host address
	host := net.ParseIP(node.HostInfo.Host)
	if host == nil {
		return nil, errors.Errorf("error parsing node host %v: is not a valid textual representation of an IP address",
			node.HostInfo.Host)
	}

	// create AltNames with defaults DNSNames/IPs
	altNames := &certutil.AltNames{
		DNSNames: []string{"localhost"},
		IPs:      []net.IP{host, net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	if cfg.Etcd.Local != nil {
		if certName == kubeadmconstants.EtcdServerCertName {
			appendSANsToAltNames(altNames, cfg.Etcd.Local.ServerCertSANs, kubeadmconstants.EtcdServerCertName)
		} else if certName == kubeadmconstants.EtcdPeerCertName {
			exAltNames := append(cfg.Etcd.Local.PeerCertSANs, node.Name)
			appendSANsToAltNames(altNames, exAltNames, kubeadmconstants.EtcdPeerCertName)
		}
	}

	if certName == kubeadmconstants.EtcdServerCertName {
		appendSANsToAltNames(altNames, []string{node.Name}, kubeadmconstants.EtcdServerCertName)
	} else if certName == kubeadmconstants.EtcdPeerCertName {
		appendSANsToAltNames(altNames, []string{node.Name}, kubeadmconstants.EtcdPeerCertName)
	}

	return altNames, nil
}
