package cert

import (
	"crypto"
	"crypto/x509"
	"github.com/yuyicai/kubei/pkg/util/pki"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"testing"
	"time"
)

func SetCfg() *kubeadmapi.InitConfiguration {
	kubeadmCfg := &kubeadmapi.InitConfiguration{}

	return kubeadmCfg
}

func TestCreateServiceAccountKeyAndPublicKeyFiles(t *testing.T) {
	type args struct {
		keyType x509.PublicKeyAlgorithm
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "SA",
			args: args{keyType: x509.RSA},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := CreateServiceAccountKeyAndPublicKey(tt.args.keyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateServiceAccountKeyAndPublicKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			key, _ := pki.EncodePrivateKeyPEM(got)
			publicKey, _ := pki.EncodePublicKeyPEM(got1)
			t.Log(string(key))
			t.Log(string(publicKey))
		})
	}
}

func TestCreateCACertAndKeyFiles(t *testing.T) {
	type args struct {
		certSpec     *KubeadmCert
		cfg          *kubeadmapi.InitConfiguration
		notAfterTime time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "CA",
			args: args{
				certSpec:     &KubeadmCertRootCA,
				cfg:          SetCfg(),
				notAfterTime: 24 * time.Hour * 365,
			},
		},

		{
			name: "CA, 20 year",
			args: args{
				certSpec:     &KubeadmCertRootCA,
				cfg:          SetCfg(),
				notAfterTime: 24 * time.Hour * 365 * 20,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := CreateCACertAndKey(tt.args.certSpec, tt.args.cfg, tt.args.notAfterTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCACertAndKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			caCert := pki.EncodeCertPEM(got)
			caKey, _ := pki.EncodePrivateKeyPEM(got1)
			t.Log(string(caCert))
			t.Log(string(caKey))
			t.Log(got.NotBefore, got.NotAfter)

		})
	}
}

func TestAllCert(t *testing.T) {
	type args struct {
		certificates Certificates
		cfg          *kubeadmapi.InitConfiguration
		notAfterTime time.Duration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "All cert, 1 year",
			args: args{
				certificates: GetDefaultCertList(),
				cfg:          SetCfg(),
				notAfterTime: 24 * time.Hour * 365,
			},
		},

		{
			name: "All cert, 20 year",
			args: args{
				certificates: GetDefaultCertList(),
				cfg:          SetCfg(),
				notAfterTime: 24 * time.Hour * 365 * 20,
			},
		},
	}
	for _, tt := range tests {

		cfg := tt.args.cfg
		cfg.LocalAPIEndpoint.AdvertiseAddress = "172.16.0.111"
		cfg.Networking.ServiceSubnet = "10.96.0.0/12"

		// etcd server peer cert
		cfg.NodeRegistration.Name = "test"

		var lastCACertCfg *KubeadmCert
		var lastCACert *x509.Certificate
		var lastCAKey crypto.Signer
		for _, c := range tt.args.certificates {
			t.Run(tt.name, func(t *testing.T) {
				t.Log("\n", c.Name)
				var cert *x509.Certificate
				var key crypto.Signer
				var err error
				if c.CAName == "" {
					cert, key, err = CreateCACertAndKey(c, cfg, tt.args.notAfterTime)
					if (err != nil) != tt.wantErr {
						t.Errorf("CreateCACertAndKey() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					lastCACertCfg = c
					lastCACert = cert
					lastCAKey = key

				} else {
					cert, key, err = CreateCertAndKeyWithCA(c, lastCACertCfg, cfg, lastCACert, lastCAKey, tt.args.notAfterTime)
					if (err != nil) != tt.wantErr {
						t.Errorf("CreateCACertAndKey() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
				}

				//t.Log(string(pki.EncodeCertPEM(cert)))
				//
				//byteKey, _ := pki.EncodePrivateKeyPEM(key)
				//t.Log(string(byteKey))
				t.Log("Subject:", cert.Subject)
				t.Log(cert.IPAddresses, cert.DNSNames)
				t.Log(cert.NotBefore, cert.NotAfter)
			})
		}
	}
}
