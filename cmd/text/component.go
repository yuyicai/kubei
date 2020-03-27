package text

import (
	"bytes"
	"github.com/lithammer/dedent"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"text/template"
)

type DocekrText interface {
	Docker(rundata.Docker) (string, error)
	RemoveDocker() string
}

type KubeText interface {
	KubeComponent(string) (string, error)
	RemoveKubeComponent() string
}

type Apt struct {
}

func (Apt) Docker(d rundata.Docker) (string, error) {
	m := map[string]interface{}{
		"Version":        d.Version,
		"CGroupDriver":   d.CGroupDriver,
		"LogDriver":      d.LogDriver,
		"LogOptsMaxSize": d.LogOptsMaxSize,
		"StorageDriver":  d.StorageDriver,
	}
	t, err := template.New("text").Parse(dedent.Dedent(`
		apt-get update -qq >/dev/null && DEBIAN_FRONTEND=noninteractive apt-get -y install apt-transport-https ca-certificates curl
		curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | apt-key add -qq - >/dev/null
		cat <<EOF | tee /etc/apt/sources.list.d/docker.list
		deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable
		EOF
		apt-get update -qq >/dev/null
		{{- if ne .Version "" }}
		DOCKER_VER=$(apt-cache madison docker-ce | awk '/{{ .Version }}/ {print$3}' | head -1)
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
		{{- if eq .CGroupDriver "systemd" }}
		  "exec-opts": ["native.cgroupdriver=systemd"],
		{{- end }}
		  "log-driver": "{{ .LogDriver }}",
		  "log-opts": {
		    "max-size": "{{ .LogOptsMaxSize }}"
		  },
		  "storage-driver": "{{ .StorageDriver }}"
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

func (Apt) Containerd(version string) (string, error) {
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
        CONTAINERD_VER=$(apt-cache madison containerd.io | awk '/{{ .version }}/ {print$3}' | head -1)
        apt-get -y install -qq containerd.io=$CONTAINERD_VER
        {{- else }}
        apt-get -y install -qq containerd.io
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

func (Yum) Docker(d rundata.Docker) (string, error) {
	m := map[string]interface{}{
		"Version":        d.Version,
		"CGroupDriver":   d.CGroupDriver,
		"LogDriver":      d.LogDriver,
		"LogOptsMaxSize": d.LogOptsMaxSize,
		"StorageDriver":  d.StorageDriver,
	}
	t, err := template.New("ver").Parse(dedent.Dedent(`
		yum install -y -q yum-utils
		yum-config-manager --add-repo \
		  https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
		{{- if ne .Version "" }}
		DOCKER_VER=$(yum list docker-ce --showduplicates | awk '/{{ .Version }}/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
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
		{{- if eq .CGroupDriver "systemd" }}
		  "exec-opts": ["native.cgroupdriver=systemd"],
		{{- end }}
		  "log-driver": "{{ .LogDriver }}",
		  "log-opts": {
		    "max-size": "{{ .LogOptsMaxSize }}"
		  },
		{{- if eq .StorageDriver "overlay2" }}
		  "storage-driver": "overlay2",
		  "storage-opts": [
		    "overlay2.override_kernel_check=true"
		  ]
		{{- end }}
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

func (Yum) Containerd(version string) (string, error) {
	m := map[string]interface{}{
		"version": version,
	}
	t, err := template.New("ver").Parse(dedent.Dedent(`
        yum install -y -q yum-utils
        yum-config-manager --add-repo \
          https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
        {{- if ne .version "" }}
        CONTAINERD_VER=$(yum list docker-ce --showduplicates | awk '/{{ .version }}/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
        yum install -y -q containerd.io=CONTAINERD_VER
        {{- else }}
        yum install -y -q containerd.io
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
	case constants.InstallationTypeApt:
		return &Apt{}
	case constants.InstallationTypeYum:
		return &Yum{}
	}
	return nil
}

func NewKubeText(installationType int) KubeText {
	switch installationType {
	case constants.InstallationTypeApt:
		return &Apt{}
	case constants.InstallationTypeYum:
		return &Yum{}
	}
	return nil
}
