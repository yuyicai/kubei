package cert

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"github.com/pkg/errors"
	"github.com/yuyicai/kubei/config/constants"
	"k8s.io/klog"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	kubeadmconstants "k8s.io/kubernetes/cmd/kubeadm/app/constants"
	"k8s.io/kubernetes/cmd/kubeadm/app/util/pkiutil"
	"time"
)

// CreateServiceAccountKeyAndPublicKey creates new public/private key files for signing service account users.
func CreateServiceAccountKeyAndPublicKey(keyType x509.PublicKeyAlgorithm) (crypto.Signer, crypto.PublicKey, error) {
	klog.V(1).Infoln("creating new public/private key files for signing service account users")

	key, err := pkiutil.NewPrivateKey(keyType)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("[certs] Generating %q key and public key\n", kubeadmconstants.ServiceAccountKeyBaseName)

	return key, key.Public(), nil
}

// CreateCACertAndKey generates and writes out a given certificate authority.
func CreateCACertAndKey(certSpec *KubeadmCert, cfg *kubeadmapi.InitConfiguration,
	notAfterTime time.Duration) (*x509.Certificate, crypto.Signer, error) {
	if certSpec.CAName != "" {
		return nil, nil, errors.Errorf("this function should only be used for CAs, but cert %s has CA %s", certSpec.Name, certSpec.CAName)
	}

	if notAfterTime < constants.DefaultCertNotAfterTime {
		notAfterTime = constants.DefaultCertNotAfterTime
	}
	certSpec.config.NotAfterTime = notAfterTime

	klog.V(1).Infof("creating a new certificate authority for %s", certSpec.Name)

	return certSpec.CreateAsCA(cfg)
}

// CreateCertAndKeyWithCA
func CreateCertAndKeyWithCA(certSpec *KubeadmCert, caCertSpec *KubeadmCert, cfg *kubeadmapi.InitConfiguration,
	caCert *x509.Certificate, caKey crypto.Signer, notAfterTime time.Duration) (*x509.Certificate, crypto.Signer, error) {
	if certSpec.CAName != caCertSpec.Name {
		return nil, nil, errors.Errorf("expected CAname for %s to be %q, but was %s", certSpec.Name, certSpec.CAName, caCertSpec.Name)
	}
	certSpec.config.NotAfterTime = notAfterTime
	return certSpec.CreateFromCA(cfg, caCert, caKey)
}
