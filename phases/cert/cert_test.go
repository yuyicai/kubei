package cert

import (
	"crypto/md5"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"github.com/yuyicai/kubei/config/rundata"
	"github.com/yuyicai/kubei/pkg/util/pki"
	kubeadmapi "k8s.io/kubernetes/cmd/kubeadm/app/apis/kubeadm"
	"strconv"
	"strings"
	"sync"
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

//func TestAllCert(t *testing.T) {
//	type args struct {
//		certificates rundata.Certificates
//		cfg          *kubeadmapi.InitConfiguration
//		notAfterTime time.Duration
//	}
//	tests := []struct {
//		name    string
//		args    args
//		wantErr bool
//	}{
//		{
//			name: "All cert, 1 year",
//			args: args{
//				certificates: rundata.GetDefaultCertList(),
//				cfg:          SetCfg(),
//				notAfterTime: 24 * time.Hour * 365,
//			},
//		},
//
//		{
//			name: "All cert, 20 year",
//			args: args{
//				certificates: rundata.GetDefaultCertList(),
//				cfg:          SetCfg(),
//				notAfterTime: 24 * time.Hour * 365 * 20,
//			},
//		},
//	}
//	for _, tt := range tests {
//
//		cfg := tt.args.cfg
//		cfg.LocalAPIEndpoint.AdvertiseAddress = "172.16.0.111"
//		cfg.Networking.ServiceSubnet = "10.96.0.0/12"
//
//		// etcd server peer cert
//		cfg.NodeRegistration.Name = "test"
//
//		var lastCACertCfg *rundata.Cert
//		var lastCACert *x509.Certificate
//		var lastCAKey crypto.Signer
//		for _, c := range tt.args.certificates {
//			t.Run(tt.name, func(t *testing.T) {
//				t.Log("\n", c.Name)
//
//				if c.CAName == "" {
//					err := CreateCACertAndKey(c, cfg, tt.args.notAfterTime)
//					if (err != nil) != tt.wantErr {
//						t.Errorf("CreateCACertAndKey() error = %v, wantErr %v", err, tt.wantErr)
//						return
//					}
//
//					lastCACertCfg = c
//					lastCACert = c.Cert
//					lastCAKey = c.Key
//
//				} else {
//					err := CreateCertAndKeyWithCA(c, lastCACertCfg, cfg, lastCACert, lastCAKey, tt.args.notAfterTime)
//					if (err != nil) != tt.wantErr {
//						t.Errorf("CreateCACertAndKey() error = %v, wantErr %v", err, tt.wantErr)
//						return
//					}
//				}
//
//				t.Log("Subject:", c.Cert.Subject)
//				t.Log(c.Cert.IPAddresses, c.Cert.DNSNames)
//				t.Log(c.Cert.NotBefore, c.Cert.NotAfter)
//			})
//		}
//	}
//}

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

	var nodes []*rundata.Node
	nodes = make([]*rundata.Node, 10)

	fmt.Printf("%+v\n", nodes)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.args.cfg
			cfg.LocalAPIEndpoint.AdvertiseAddress = "172.16.0.111"
			cfg.Networking.ServiceSubnet = "10.96.0.0/12"

			// etcd server peer cert
			cfg.NodeRegistration.Name = "test"
			//certTree := rundata.CertificateTree{}
			tt.args.node.Name = "yyyy"

			nodes[0] = &rundata.Node{}
			nodes[0].Name = "yyzz"

			if err := CreatePKIAssets(nodes[0], tt.args.cfg, tt.args.notAfterTime, certTree); (err != nil) != tt.wantErr {
				t.Errorf("CreatePKIAssets() error = %v, wantErr %v", err, tt.wantErr)
			}

			certTree = nodes[0].CertificateTree

			for ca, certs := range certTree {
				fmt.Println(ca.BaseName)
				fmt.Printf("CA %s.crt not after time: %v\n", ca.Name, ca.Cert.NotAfter)
				for _, cert := range certs {
					fmt.Printf("Cert %s.crt not after time: %v\n", cert.Name, cert.Cert.NotAfter)
					fmt.Printf("%+v\n", cert.Cert.Subject)
				}
				fmt.Println("---------------------------------------------")
			}
			fmt.Println("2345", nodes[0].CertificateTree)

		})
	}
}

func TestCreatePKIAssetsLock(t *testing.T) {
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

	var nodes []*rundata.Node
	nodes = make([]*rundata.Node, 10)

	//fmt.Printf("%+v\n", nodes)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.args.cfg
			cfg.LocalAPIEndpoint.AdvertiseAddress = "172.16.0.111"
			cfg.Networking.ServiceSubnet = "10.96.0.0/12"

			// etcd server peer cert
			cfg.NodeRegistration.Name = "test"
			//certTree := rundata.CertificateTree{}
			tt.args.node.Name = "yyyy"

			wg := sync.WaitGroup{}

			for i, n := range nodes {
				wg.Add(1)
				i := i
				n := n
				go func() {
					defer wg.Done()
					fmt.Println("---------------------------------------------")
					n = &rundata.Node{}
					n.Name = "yyzz" + strconv.Itoa(i)

					if err := CreatePKIAssets(n, tt.args.cfg, tt.args.notAfterTime, certTree); (err != nil) != tt.wantErr {
						t.Errorf("CreatePKIAssets() error = %v, wantErr %v", err, tt.wantErr)
					}

					certTree = n.CertificateTree

					for _, certs := range certTree {
						//fmt.Println(ca.BaseName)
						//fmt.Printf("CA %s.crt not after time: %v\n", ca.Name, ca.Cert.NotAfter)
						for _, cert := range certs {
							if strings.Contains(cert.Name, "etcd") {
								fmt.Printf("Cert %s.crt not after time: %v\n", cert.Name, cert.Cert.NotAfter)
								fmt.Printf("%+v\n", cert.Cert.Subject)
							}

						}
					}
				}()

				wg.Wait()

			}

		})
	}
}

func TestCreateCert(t *testing.T) {
	type args struct {
		c *rundata.Cluster
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create cert",
			args: args{c: rundata.NewCluster()},
		},
	}
	for _, tt := range tests {

		tt.args.c.ClusterNodes.Masters = make([]*rundata.Node, 10)
		tt.args.c.Kubeadm.LocalAPIEndpoint.AdvertiseAddress = "172.16.0.111"
		tt.args.c.Kubeadm.Networking.ServiceSubnet = "10.96.0.0/12"
		tt.args.c.Kubeadm.LocalAPIEndpoint.BindPort = 6443

		for i, _ := range tt.args.c.ClusterNodes.Masters {
			tt.args.c.ClusterNodes.Masters[i] = &rundata.Node{
				CertificateTree: rundata.CertificateTree{},
			}
			tt.args.c.ClusterNodes.Masters[i].Name = "node" + strconv.Itoa(i)
			tt.args.c.ClusterNodes.Masters[i].HostInfo.Host = "192.168.0." + strconv.Itoa(i)
		}

		//fmt.Printf("%+v", tt.args.c.ClusterNodes.Masters)

		if err := CreateCert(tt.args.c); (err != nil) != tt.wantErr {
			t.Errorf("CreateCert() error = %v, wantErr %v", err, tt.wantErr)
		}

		for _, node := range tt.args.c.ClusterNodes.Masters {
			//fmt.Println(tt.args.c.ClusterNodes.Masters[i1].CertificateTree)
			fmt.Println("+++++++++", node.Name)
			for _, certTree := range node.CertificateTree {
				//fmt.Println(ca.BaseName)
				//fmt.Printf("CA %s.crt not after time: %v\n", ca.Name, ca.Cert.NotAfter)
				for _, cert := range certTree {
					//if strings.Contains(cert.Name, "etcd-s") {
					//	fmt.Printf("Cert %s.crt not after time: %v\n", cert.Name, cert.Cert.NotAfter)
					//	fmt.Printf("Subject: %+v\n", cert.Cert.Subject)
					//	md5Ctx := md5.New()
					//	md5Ctx.Write(cert.Cert.Raw)
					//	fmt.Println(hex.EncodeToString(md5Ctx.Sum(nil)))
					//}
					fmt.Printf("Cert %s.crt not after time: %v\n", cert.Name, cert.Cert.NotAfter)
					fmt.Printf("Subject: %+v\n", cert.Cert.Subject)
					md5Ctx := md5.New()
					md5Ctx.Write(cert.Cert.Raw)
					fmt.Println(hex.EncodeToString(md5Ctx.Sum(nil)))

				}
			}
			fmt.Println("+++++++++", node.Name)
			fmt.Println("")
		}
	}
}

func TestAllCreateCert(t *testing.T) {
	type args struct {
		c *rundata.Cluster
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "create cert",
			args: args{c: rundata.NewCluster()},
		},
	}
	for _, tt := range tests {

		tt.args.c.ClusterNodes.Masters = make([]*rundata.Node, 10)
		tt.args.c.Kubeadm.LocalAPIEndpoint.AdvertiseAddress = "172.16.0.111"
		tt.args.c.Kubeadm.Networking.ServiceSubnet = "10.96.0.0/12"
		tt.args.c.Kubeadm.LocalAPIEndpoint.BindPort = 6443

		for i, _ := range tt.args.c.ClusterNodes.Masters {
			tt.args.c.ClusterNodes.Masters[i] = &rundata.Node{
				CertificateTree: rundata.CertificateTree{},
			}
			tt.args.c.ClusterNodes.Masters[i].Name = "node" + strconv.Itoa(i)
			tt.args.c.ClusterNodes.Masters[i].HostInfo.Host = "192.168.0." + strconv.Itoa(i)
		}

		if err := CreateCert(tt.args.c); (err != nil) != tt.wantErr {
			t.Errorf("CreateCert() error = %v, wantErr %v", err, tt.wantErr)
		}

		for _, node := range tt.args.c.ClusterNodes.Masters {
			fmt.Println("+++++++++", node.Name)
			for ca, certs := range node.CertificateTree {
				fmt.Printf("CA Subject: %v\n", ca.Cert.Subject)
				for _, cert := range certs {
					fmt.Printf("Cert Name: %s, Cert Subject: %v, not after time: %v\n", cert.Name, cert.Cert.Subject, cert.Cert.NotAfter)
					fmt.Println("IP:", cert.Cert.IPAddresses, "DNS:", cert.Cert.DNSNames)

					if cert.IsKubeConfig {
						config, _ := EncodeKubeConfig(cert.KubeConfig)
						fmt.Println(string(config))
					}
				}
			}
			fmt.Println("+++++++++", node.Name)
			fmt.Println("")
		}
	}
}
