package cert

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"

	"github.com/fatih/color"
	"k8s.io/apimachinery/pkg/runtime"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"

	"github.com/yuyicai/kubei/internal/operator"
	"github.com/yuyicai/kubei/internal/rundata"
	"github.com/yuyicai/kubei/pkg/pki"
)

func SendCert(c *rundata.Cluster) error {
	encodedPrivatKey, encodedPublicKey, err := CreateEncodeServiceAccountKeyAndPublicKey(x509.RSA)
	if err != nil {
		return err
	}

	encodedPrivatKeyBase64 := base64.StdEncoding.EncodeToString(encodedPrivatKey)
	encodedPublicKeyBase64 := base64.StdEncoding.EncodeToString(encodedPublicKey)

	return operator.RunOnMasters(c, func(node *rundata.Node, c *rundata.Cluster) error {
		if err := sendServiceAccountKeyAndPublicKey(node, encodedPrivatKeyBase64, encodedPublicKeyBase64); err != nil {
			return err
		}

		if err := sendCertAndKubeConfig(node); err != nil {
			return err
		}

		fmt.Printf("[%s] [cert] create certificates: %s\n", node.HostInfo.Host, color.HiGreenString("done✅️"))
		return nil
	})

}

func sendCertAndKubeConfig(node *rundata.Node) error {

	if err := node.Run("mkdir -p /etc/kubernetes/pki/etcd"); err != nil {
		return err
	}

	certTree := node.CertificateTree

	for ca, certs := range certTree {
		if err := sendCert(node, ca); err != nil {
			return err
		}

		for _, cert := range certs {
			if cert.IsKubeConfig {
				if err := sendKubeConfig(node, cert); err != nil {
					return err
				}
				continue
			}
			if err := sendCert(node, cert); err != nil {
				return err
			}
		}
	}

	return nil
}

func sendCert(node *rundata.Node, c *rundata.Cert) error {
	// TODO set umask
	// send cert
	encodeCert := pki.EncodeCertPEM(c.Cert)
	encodeCertBase64 := base64.StdEncoding.EncodeToString(encodeCert)
	if err := node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/pki/%s.crt", encodeCertBase64, c.BaseName)); err != nil {
		return err
	}

	// send key
	encodedKey, err := pki.EncodePrivateKeyPEM(c.Key)
	if err != nil {
		return err
	}
	encodedKeyBase64 := base64.StdEncoding.EncodeToString(encodedKey)
	return node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/pki/%s.key", encodedKeyBase64, c.BaseName))
}

func sendKubeConfig(node *rundata.Node, c *rundata.Cert) error {
	encodedKubeConfig, err := EncodeKubeConfig(c.KubeConfig)
	if err != nil {
		return err
	}
	encodedKubeConfigBase64 := base64.StdEncoding.EncodeToString(encodedKubeConfig)

	//if c.Name == "admin" {
	//	if err := node.Run(fmt.Sprintf("mkdir -p $HOME/.kube && echo %s | base64 -d > $HOME/.kube/config", encodedKubeConfigBase64)); err != nil {
	//		return err
	//	}
	//}

	return node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/%s", encodedKubeConfigBase64, c.BaseName))
}

func sendServiceAccountKeyAndPublicKey(node *rundata.Node, privatKey, publicKey string) error {
	if err := node.Run("mkdir -p /etc/kubernetes/pki/etcd"); err != nil {
		return err
	}

	if err := node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/pki/%s", privatKey, "sa.key")); err != nil {
		return err
	}
	return node.Run(fmt.Sprintf("echo %s | base64 -d > /etc/kubernetes/pki/%s", publicKey, "sa.pub"))
}

// EncodeKubeConfig serializes the config to yaml.
// Encapsulates serialization without assuming the destination is a file.
func EncodeKubeConfig(config clientcmdapi.Config) ([]byte, error) {
	return runtime.Encode(clientcmdlatest.Codec, &config)
}
