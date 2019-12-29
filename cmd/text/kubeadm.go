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
        kubeadm join {{ .controlPlaneEndpoint }} --token {{ .token }}  \
          --discovery-token-ca-cert-hash sha256:{{ .caCertHash }} \
          --node-name {{ .nodeName }} \
          --ignore-preflight-errors=DirAvailable--etc-kubernetes-manifests
	`))
	if err != nil {
		return "", err
	}

	_, err = t.New(JoinControlPlane).Parse(dedent.Dedent(`
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
        mkdir -p $HOME/.kube
        yes | cp /etc/kubernetes/admin.conf $HOME/.kube/config
	`)
}

func ChownKubectlConfig() string {
	return "chown $SUDO_USER:$SUDO_UID $HOME/.kube/config"
}

func SwapOff() string {
	return dedent.Dedent(`
        swapoff -a && sysctl -w vm.swappiness=0
        sed -i "/swap/ s/^#*/#/" /etc/fstab
	`)
}

func Iptables() string {
	cmd := dedent.Dedent(`
        cat <<EOF | tee /etc/sysctl.d/99-k8s-sysctl.conf 
        net.ipv4.ip_forward=1
        net.bridge.bridge-nf-call-iptables=1
        net.bridge.bridge-nf-call-arptables=1
        net.bridge.bridge-nf-call-ip6tables=1
        EOF
        sysctl --system
	`)
	return cmd
}
