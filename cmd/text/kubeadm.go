package text

import (
	"bytes"
	"github.com/lithammer/dedent"
	"text/template"

	"github.com/yuyicai/kubei/config/rundata"
)

const (
	Init             = "init"
	JoinNode         = "joinNode"
	JoinControlPlane = "joinControlPlane"
)

func Kubeadm(tmplName, nodeName string, kubeadmCfg *rundata.Kubeadm) (string, error) {
	token := kubeadmCfg.Token
	m := map[string]interface{}{
		"nodeName":             nodeName,
		"imageRepository":      kubeadmCfg.ImageRepository,
		"podNetworkCidr":       kubeadmCfg.Networking.PodSubnet,
		"serviceCidr":          kubeadmCfg.Networking.ServiceSubnet,
		"controlPlaneEndpoint": kubeadmCfg.ControlPlaneEndpoint,
		"token":                token.Token,
		"caCertHash":           token.CaCertHash,
		"certificateKey":       token.CertificateKey,
	}

	t, err := template.New(Init).Parse(dedent.Dedent(`
        swapoff -a
        kubeadm init \
          --image-repository {{ .imageRepository }} \
          --pod-network-cidr {{ .podNetworkCidr }} \
          --service-cidr {{ .serviceCidr }} \
          --upload-certs \
          --control-plane-endpoint {{ .controlPlaneEndpoint }} \
          --node-name {{ .nodeName }}
	`))
	if err != nil {
		return "", err
	}

	_, err = t.New(JoinNode).Parse(dedent.Dedent(`
        swapoff -a
        kubeadm join {{ .controlPlaneEndpoint }} --token {{ .token }}  \
          --discovery-token-ca-cert-hash sha256:{{ .caCertHash }} \
          --node-name {{ .nodeName }} \
          --ignore-preflight-errors=DirAvailable--etc-kubernetes-manifests
	`))
	if err != nil {
		return "", err
	}

	_, err = t.New(JoinControlPlane).Parse(dedent.Dedent(`
        swapoff -a
        yes | kubeadm reset
        kubeadm join {{ .controlPlaneEndpoint }} \
          --token {{ .token }} \
          --discovery-token-ca-cert-hash sha256:{{ .caCertHash }} \
          --certificate-key {{ .certificateKey }} \
          --control-plane \
          --node-name {{ .nodeName }}
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	err = t.ExecuteTemplate(&cmdBuff, tmplName, m)
	if err != nil {
		return "", err
	}

	cmd := cmdBuff.String()
	return cmd, nil
}

func CopyAdminConfig() string {
	return dedent.Dedent(`
        sudo mkdir $HOME/.kube
        yes | sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
        sudo chown $(id -u):$(id -g) $HOME/.kube/config
	`)
}

func SwapOff() string {
	return dedent.Dedent(`
        sudo mkdir $HOME/.kube
        yes | sudo cp /etc/kubernetes/admin.conf $HOME/.kube/config
        sudo chown $(id -u):$(id -g) $HOME/.kube/config
	`)
}
