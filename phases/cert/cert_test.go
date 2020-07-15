package cert

import (
	"crypto"
	"crypto/x509"
	"fmt"
	"github.com/yuyicai/kubei/config/rundata"
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

func TestAllCert(t *testing.T) {
	type args struct {
		certificates rundata.Certificates
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
				certificates: rundata.GetDefaultCertList(),
				cfg:          SetCfg(),
				notAfterTime: 24 * time.Hour * 365,
			},
		},

		{
			name: "All cert, 20 year",
			args: args{
				certificates: rundata.GetDefaultCertList(),
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

		var lastCACertCfg *rundata.Cert
		var lastCACert *x509.Certificate
		var lastCAKey crypto.Signer
		for _, c := range tt.args.certificates {
			t.Run(tt.name, func(t *testing.T) {
				t.Log("\n", c.Name)

				if c.CAName == "" {
					err := CreateCACertAndKey(c, cfg, tt.args.notAfterTime)
					if (err != nil) != tt.wantErr {
						t.Errorf("CreateCACertAndKey() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					lastCACertCfg = c
					lastCACert = c.Cert
					lastCAKey = c.Key

				} else {
					err := CreateCertAndKeyWithCA(c, lastCACertCfg, cfg, lastCACert, lastCAKey, tt.args.notAfterTime)
					if (err != nil) != tt.wantErr {
						t.Errorf("CreateCACertAndKey() error = %v, wantErr %v", err, tt.wantErr)
						return
					}
				}

				t.Log("Subject:", c.Cert.Subject)
				t.Log(c.Cert.IPAddresses, c.Cert.DNSNames)
				t.Log(c.Cert.NotBefore, c.Cert.NotAfter)
			})
		}
	}
}

func TestCreatePKIAssets(t *testing.T) {
	type args struct {
		cfg          *kubeadmapi.InitConfiguration
		notAfterTime time.Duration
		node         *rundata.Node
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "PKI Assets, 1 year",
			args: args{
				cfg:          SetCfg(),
				notAfterTime: 24 * time.Hour * 365,
				node: &rundata.Node{
					CertificateTree: rundata.CertificateTree{},
				},
			},
		},
		{
			name: "PKI Assets, 20 year",
			args: args{
				cfg:          SetCfg(),
				notAfterTime: 24 * time.Hour * 365 * 20,
				node: &rundata.Node{
					CertificateTree: rundata.CertificateTree{},
				},
			},
		},
	}

	certTree := rundata.CertificateTree{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.args.cfg
			cfg.LocalAPIEndpoint.AdvertiseAddress = "172.16.0.111"
			cfg.Networking.ServiceSubnet = "10.96.0.0/12"

			// etcd server peer cert
			cfg.NodeRegistration.Name = "test"

			//certTree := rundata.CertificateTree{}

			if err := CreatePKIAssets(tt.args.node, tt.args.cfg, tt.args.notAfterTime, certTree); (err != nil) != tt.wantErr {
				t.Errorf("CreatePKIAssets() error = %v, wantErr %v", err, tt.wantErr)
			}

			certTree = tt.args.node.CertificateTree

			for ca, certs := range certTree {
				fmt.Println(ca.BaseName)
				fmt.Printf("CA %s.crt not after time: %v\n", ca.Name, ca.Cert.NotAfter)
				for _, cert := range certs {
					fmt.Printf("CA %s.crt not after time: %v\n", cert.Name, cert.Cert.NotAfter)
					//fmt.Printf("%+v\n", cert.Cert)
				}
				fmt.Println("---------------------------------------------")
			}

		})
	}
}
