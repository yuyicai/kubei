package text

import (
	"bytes"
	"github.com/lithammer/dedent"
	"github.com/yuyicai/kubei/config/rundata"
	"text/template"
)

type DocekrText interface {
	Docker(string) (string, error)
	RemoveDocker() string
}

type KubeText interface {
	KubeComponent(string) (string, error)
	RemoveKubeComponent() string
}

type Apt struct {
}

func (Apt) Docker(version string) (string, error) {
	m := map[string]interface{}{
		"version": version,
	}
	t, err := template.New("ver").Parse(dedent.Dedent(`
        apt-get update -qq >/dev/null && DEBIAN_FRONTEND=noninteractive apt-get -y install apt-transport-https ca-certificates curl
        curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | apt-key add -qq - >/dev/null
        cat <<EOF | tee /etc/apt/sources.list.d/docker.list
        deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable
        EOF
        apt-get update -qq >/dev/null
        {{- if ne .version "" }}
        DOCKER_VER=$(apt-cache madison docker-ce | awk '/{{ .version }}/ {print$3}' | head -1)
        apt-get -y install -qq docker-ce=$DOCKER_VER docker-ce-cli=$DOCKER_VER containerd.io
        {{- else }}
        apt-get -y install -qq docker-ce docker-ce-cli containerd.io
        {{- end }}
        cat <<EOF | tee /etc/docker/daemon.json
        {
          "registry-mirrors": [
              "https://dockerhub.azk8s.cn",
              "https://hub-mirror.c.163.com"
          ],
          "exec-opts": ["native.cgroupdriver=systemd"],
          "log-driver": "json-file",
          "log-opts": {
            "max-size": "500m"
          },
          "storage-driver": "overlay2"
        }
        EOF
        mkdir -p /etc/systemd/system/docker.service.d
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.Execute(&cmdBuff, m); err != nil {
		return "", err
	}

	cmd := cmdBuff.String()
	return cmd, nil
}

func (Apt) KubeComponent(version string) (string, error) {
	m := map[string]interface{}{
		"version": version,
	}
	t, err := template.New("ver").Parse(dedent.Dedent(`
        apt-get update -qq && apt-get install -qq -y apt-transport-https curl
        curl -s https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -
        cat <<EOF | tee /etc/apt/sources.list.d/kubernetes.list
        deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
        EOF
        apt-get update -qq
        {{- if ne .version "" }}
        KUBE_VER=$(apt-cache madison kubelet | awk '/{{ .version }}/ {print$3}' | head -1)
        apt-get install -qq -y --allow-change-held-packages kubelet=$KUBE_VER kubeadm=$KUBE_VER kubectl=$KUBE_VER
        {{- else }}
        apt-get install -qq -y --allow-change-held-packages kubelet kubeadm kubectl
        {{- end }}
        apt-mark hold kubelet kubeadm kubectl
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.Execute(&cmdBuff, m); err != nil {
		return "", err
	}

	cmd := cmdBuff.String()
	return cmd, nil
}

func (Apt) RemoveDocker() string {
	return "apt-get remove -y docker-ce docker-ce-cli containerd.io || true"
}

func (Apt) RemoveKubeComponent() string {
	return "apt-get remove -y --allow-change-held-packages kubelet kubeadm kubectl || true"
}

type Yum struct {
}

func (Yum) Docker(version string) (string, error) {
	m := map[string]interface{}{
		"version": version,
	}
	t, err := template.New("ver").Parse(dedent.Dedent(`
        yum install -y -q yum-utils
        yum-config-manager --add-repo \
          https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
        {{- if ne .version "" }}
        DOCKER_VER=$(yum list docker-ce --showduplicates | awk '/{{ .version }}/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
        yum install -y -q docker-ce-$DOCKER_VER docker-ce-cli-$DOCKER_VER containerd.io
        {{- else }}
        yum install -y -q docker-ce docker-ce-cli containerd.io
        {{- end }}
        mkdir -p /etc/docker
        cat <<EOF | tee /etc/docker/daemon.json
        {
          "registry-mirrors": [
              "https://dockerhub.azk8s.cn",
              "https://hub-mirror.c.163.com"
          ],
          "exec-opts": ["native.cgroupdriver=systemd"],
          "log-driver": "json-file",
          "log-opts": {
            "max-size": "100m"
          },
          "storage-driver": "overlay2",
          "storage-opts": [
            "overlay2.override_kernel_check=true"
          ]
        }
        EOF
        mkdir -p /etc/systemd/system/docker.service.d
    `))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.Execute(&cmdBuff, m); err != nil {
		return "", err
	}

	cmd := cmdBuff.String()
	return cmd, nil
}

func (Yum) KubeComponent(version string) (string, error) {
	m := map[string]interface{}{
		"version": version,
	}
	t, err := template.New("ver").Parse(dedent.Dedent(`
        cat <<EOF | tee /etc/yum.repos.d/kubernetes.repo
        [kubernetes]
        name=Kubernetes
        baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
        enabled=1
        gpgcheck=1
        repo_gpgcheck=1
        gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
        EOF
        setenforce 0
        sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
        {{- if ne .version "" }}
        KUBE_VER=$(yum list kubelet --showduplicates | awk '/{{ .version }}/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
        yum install -y kubelet-$KUBE_VER kubeadm-$KUBE_VER kubectl-$KUBE_VER --disableexcludes=kubernetes
        {{- else }}
        yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
        {{- end }}
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.Execute(&cmdBuff, m); err != nil {
		return "", err
	}

	cmd := cmdBuff.String()
	return cmd, nil
}

func (Yum) RemoveDocker() string {
	return "yum remove -y docker-ce docker-ce-cli containerd.io || true"
}

func (Yum) RemoveKubeComponent() string {
	return "yum remove -y kubelet kubeadm kubectl  || true"
}

func NewContainerEngineText(installationType int) DocekrText {
	switch installationType {
	case rundata.Apt:
		return &Apt{}
	case rundata.Yum:
		return &Yum{}
	}
	return nil
}

func NewKubeText(installationType int) KubeText {
	switch installationType {
	case rundata.Apt:
		return &Apt{}
	case rundata.Yum:
		return &Yum{}
	}
	return nil
}
