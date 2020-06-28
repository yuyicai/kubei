package tmpl

import (
	"bytes"
	"github.com/lithammer/dedent"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"text/template"
)

type DocekrText interface {
	Docker(installTyped string, dockerData rundata.Docker) (string, error)
	RemoveDocker() string
}

type KubeText interface {
	KubeComponent(version, installType string) (string, error)
	RemoveKubeComponent() string
}

type Apt struct {
}

func (Apt) Docker(installTyped string, d rundata.Docker) (string, error) {
	m := map[string]interface{}{
		"version":        d.Version,
		"cgroupDriver":   d.CGroupDriver,
		"logDriver":      d.LogDriver,
		"logOptsMaxSize": d.LogOptsMaxSize,
		"storageDriver":  d.StorageDriver,
	}
	t, err := template.New("text").Parse(dedent.Dedent(`
		{{ define "config" }}
		mkdir -p /etc/docker/ || true
		cat <<EOF | tee /etc/docker/daemon.json
		{
		  "registry-mirrors": [
		      "https://dockerhub.mirrors.nwafu.edu.cn/",
		      "https://hub-mirror.c.163.com"
		  ],
		{{- if eq .cgroupDriver "systemd" }}
		  "exec-opts": ["native.cgroupdriver=systemd"],
		{{- end }}
		  "log-driver": "{{ .logDriver }}",
		  "log-opts": {
		    "max-size": "{{ .logOptsMaxSize }}"
		  },
		  "storage-driver": "{{ .storageDriver }}"
		}
		EOF
		mkdir -p /etc/systemd/system/docker.service.d || true
		{{ end }}
		{{ define "online" }}
		apt-get update -qq >/dev/null && DEBIAN_FRONTEND=noninteractive apt-get -y install -qq apt-transport-https ca-certificates curl
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
		{{- template "config" . -}}
		{{ end }}
		{{ define "offline" }}
		{{- template "config" . -}}
		sh /tmp/.kubei/container_engine/default.sh
		{{ end }}
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.ExecuteTemplate(&cmdBuff, installTyped, m); err != nil {
		return "", err
	}

	return cmdBuff.String(), nil
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

func (Apt) KubeComponent(version, installType string) (string, error) {
	m := map[string]interface{}{
		"version": version,
	}
	t, err := template.New("text").Parse(dedent.Dedent(`
		{{ define "online" }}
		apt-get update -qq && apt-get install -qq -y apt-transport-https curl
		curl -s https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add - >/dev/null
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
		{{ end }}
		{{ define "offline" }}
		sh /tmp/.kubei/kubernetes/default.sh
		{{ end }}
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.ExecuteTemplate(&cmdBuff, installType, m); err != nil {
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

func (Yum) Docker(installType string, d rundata.Docker) (string, error) {
	m := map[string]interface{}{
		"version":        d.Version,
		"cgroupDriver":   d.CGroupDriver,
		"logDriver":      d.LogDriver,
		"logOptsMaxSize": d.LogOptsMaxSize,
		"storageDriver":  d.StorageDriver,
	}
	t, err := template.New("text").Parse(dedent.Dedent(`
		{{ define "config" }}
		mkdir -p /etc/docker/ || true
		cat <<EOF | tee /etc/docker/daemon.json
		{
		  "registry-mirrors": [
		      "https://dockerhub.mirrors.nwafu.edu.cn/",
		      "https://hub-mirror.c.163.com"
		  ],
		{{- if eq .cgroupDriver "systemd" }}
		  "exec-opts": ["native.cgroupdriver=systemd"],
		{{- end }}
		  "log-driver": "{{ .logDriver }}",
		  "log-opts": {
		    "max-size": "{{ .logOptsMaxSize }}"
		  },
		{{- if eq .storageDriver "overlay2" }}
		  "storage-driver": "overlay2",
		  "storage-opts": [
		    "overlay2.override_kernel_check=true"
		  ]
		{{- end }}
		}
		EOF
		mkdir -p /etc/systemd/system/docker.service.d
		{{ end }}
		{{ define "online" }}
		yum install -y -q yum-utils
		yum-config-manager --add-repo \
		  https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
		{{- if ne .version "" }}
		DOCKER_VER=$(yum list docker-ce --showduplicates | awk '/{{ .version }}/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
		yum install -y -q docker-ce-$DOCKER_VER docker-ce-cli-$DOCKER_VER containerd.io
		{{- else }}
		yum install -y -q docker-ce docker-ce-cli containerd.io
		{{- end }}
		{{- template "config" . -}}
		{{ end }}
		{{ define "offline" }}
		{{- template "config" . -}}
		sh /tmp/.kubei/container_engine/default.sh
		{{ end }}
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.ExecuteTemplate(&cmdBuff, installType, m); err != nil {
		return "", err
	}

	return cmdBuff.String(), nil
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

func (Yum) KubeComponent(version, installType string) (string, error) {
	m := map[string]interface{}{
		"version": version,
	}
	t, err := template.New("ver").Parse(dedent.Dedent(`
		{{ define "selinux" }}
		setenforce 0 || true
		sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
		{{- end }}
		{{ define "online" }}
		cat <<EOF | tee /etc/yum.repos.d/kubernetes.repo
		[kubernetes]
		name=Kubernetes
		baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
		enabled=1
		gpgcheck=1
		repo_gpgcheck=1
		gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
		EOF
		{{- template "selinux" . -}}
		{{- if ne .version "" }}
		KUBE_VER=$(yum list kubelet --showduplicates | awk '/{{ .version }}/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
		yum install -y kubelet-$KUBE_VER kubeadm-$KUBE_VER kubectl-$KUBE_VER --disableexcludes=kubernetes
		{{- else }}
		yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
		{{- end }}
		{{ end }}
		{{ define "offline" }}
		{{- template "selinux" . }}
		sh /tmp/.kubei/kubernetes/default.sh
		{{ end }}
	`))
	if err != nil {
		return "", err
	}

	var cmdBuff bytes.Buffer
	if err := t.ExecuteTemplate(&cmdBuff, installType, m); err != nil {
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

func NewContainerEngineText(installationType string) DocekrText {
	switch installationType {
	case constants.PackageManagementTypeApt:
		return &Apt{}
	case constants.PackageManagementTypeYum:
		return &Yum{}
	}
	return nil
}

func NewKubeText(installationType string) KubeText {
	switch installationType {
	case constants.PackageManagementTypeApt:
		return &Apt{}
	case constants.PackageManagementTypeYum:
		return &Yum{}
	}
	return nil
}
