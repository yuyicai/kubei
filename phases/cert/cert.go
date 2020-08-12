package cert

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/pkg/errors"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	kubeadmpkiutil "k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"

	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/pkg/pki"
)

func CreateCert(c *rundata.Cluster) error {

	color.HiBlue("Creating certificates for kubernetes and etcd 📘")

	certTree := rundata.CertificateTree{}

	certNotAfterTime := constants.Year * time.Duration(c.CertNotAfterTime)

	if err := c.RunOnFirstMaster(func(node *rundata.Node) error {
		klog.V(2).Infof("[%s] [cert] Creating certificate", node.HostInfo.Host)
		klog.V(3).Infof("[%s] [cert] The cert not after time is %v", node.HostInfo.Host, certNotAfterTime)
		c.Kubeadm.NodeRegistration.Name = node.Name
		if err := CreatePKIAssets(node, &c.Kubeadm.InitConfiguration, certNotAfterTime, certTree); err != nil {
			return err
		}
		certTree = node.CertificateTree

		return node.CertificateTree.CreateKubeConfig(&c.Kubeadm.InitConfiguration)

	}); err != nil {
		return err
	}

	return c.RunOnOtherMasters(func(node *rundata.Node) error {
		klog.V(2).Infof("[%s] [cert] Creating certificate", node.HostInfo.Host)
		klog.V(3).Infof("[%s] [cert] The cert not after time is %v", node.HostInfo.Host)
		//c.Mutex.Lock()
		//c.Kubeadm.NodeRegistration.Name = node.Name

		if err := CreatePKIAssets(node, &c.Kubeadm.InitConfiguration, certNotAfterTime, certTree); err != nil {
			return err
		}
		//c.Mutex.Unlock()

		return node.CertificateTree.CreateKubeConfig(&c.Kubeadm.InitConfiguration)

	})
}

// CreatePKIAssets will create all PKI assets necessary.
func CreatePKIAssets(node *rundata.Node, cfg *kubeadmapi.InitConfiguration, notAfterTime time.Duration, certTree rundata.CertificateTree) error {
	klog.V(3).Infoln("creating PKI assets")

	var certList rundata.Certificates
	var err error

	certList = rundata.GetDefaultCertList()

	certMap := certList.AsMap()

	for cert := range certTree {
		if cert.CAName == "" {
			certMap[cert.Name] = cert
		}
	}

	node.CertificateTree, err = certMap.CertTree()
	if err != nil {
		return err
	}

	if err := node.CertificateTree.Create(node, cfg, notAfterTime); err != nil {
		return errors.Wrap(err, fmt.Sprintf("error creating PKI assets on %s", node.HostInfo.Host))
	}

	return nil
}

// CreateServiceAccountKeyAndPublicKey creates new public/private key files for signing service account users.
func CreateServiceAccountKeyAndPublicKey(keyType x509.PublicKeyAlgorithm) (crypto.Signer, crypto.PublicKey, error) {
	klog.V(3).Infoln("creating new public/private key files for signing service account users")

	key, err := kubeadmpkiutil.NewPrivateKey(keyType)
	if err != nil {
		return nil, nil, err
	}

	klog.V(3).Infoln("[certs] Generating %q key and public key", kubeadmconstants.ServiceAccountKeyBaseName)

	return key, key.Public(), nil
}

func CreateEncodeServiceAccountKeyAndPublicKey(keyType x509.PublicKeyAlgorithm) (encodedPrivatKey, encodedPublicKey []byte, err error) {
	privatKey, publicKey, err := CreateServiceAccountKeyAndPublicKey(keyType)
	if err != nil {
		return nil, nil, err
	}
	encodedPrivatKey, err = pki.EncodePrivateKeyPEM(privatKey)
	if err != nil {
		return nil, nil, err
	}
	encodedPublicKey, err = pki.EncodePublicKeyPEM(publicKey)
	if err != nil {
		return nil, nil, err
	}
	return
}

//// CreateCACertAndKey generates and writes out a given certificate authority.
//func CreateCACertAndKey(certSpec *rundata.Cert, cfg *kubeadmapi.InitConfiguration, notAfterTime time.Duration) error {
//	if certSpec.CAName != "" {
//		return errors.Errorf("this function should only be used for CAs, but cert %s has CA %s", certSpec.Name, certSpec.CAName)
//	}
//
//	if notAfterTime < constants.DefaultCertNotAfterTime {
//		notAfterTime = constants.DefaultCertNotAfterTime
//	}
//	certSpec.Config.NotAfterTime = notAfterTime
//
//	klog.V(1).Infof("creating a new certificate authority for %s", certSpec.Name)
//
//	return certSpec.CreateAsCA(cfg)
//}
//
//// CreateCertAndKeyWithCA
//func CreateCertAndKeyWithCA(certSpec *rundata.Cert, caCertSpec *rundata.Cert, cfg *kubeadmapi.InitConfiguration,
//	caCert *x509.Certificate, caKey crypto.Signer, notAfterTime time.Duration) error {
//	if certSpec.CAName != caCertSpec.Name {
//		return errors.Errorf("expected CAname for %s to be %q, but was %s", certSpec.Name, certSpec.CAName, caCertSpec.Name)
//	}
//	certSpec.Config.NotAfterTime = notAfterTime
//	return certSpec.CreateFromCA(cfg, caCert, caKey)
//}
