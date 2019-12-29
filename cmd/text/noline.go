package text

import (
	"github.com/lithammer/dedent"
	"github.com/yuyicai/kubei/config/rundata"
)

type DocekrText interface {
	Docker() string
}

type KubeText interface {
	KubeComponent() string
}

type Apt struct {
}

func (Apt) Docker() string {
	cmd := dedent.Dedent(`
        apt update && apt -y install apt-transport-https ca-certificates curl software-properties-common
        curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | apt-key add -
        add-apt-repository "deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable"
        apt update && apt -y install docker-ce docker-ce-cli containerd.io
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
	`)
	return cmd
}

func (Apt) KubeComponent() string {
	cmd := dedent.Dedent(`
        apt update && apt install -y apt-transport-https curl
        curl -s https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add -
        cat <<EOF | tee /etc/apt/sources.list.d/kubernetes.list
        deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
        EOF
        apt update
        apt install -y --allow-change-held-packages kubelet kubeadm kubectl
        apt-mark hold kubelet kubeadm kubectl
	`)
	return cmd
}

type Yum struct {
}

func (Yum) Docker() string {
	cmd := dedent.Dedent(`
        yum install -y yum-utils device-mapper-persistent-data lvm2
        yum-config-manager --add-repo \
          https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
        yum install -y docker-ce docker-ce-cli containerd.io
        mkdir /etc/docker
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
    `)
	return cmd
}

func (Yum) KubeComponent() string {
	cmd := dedent.Dedent(`
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
        yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
	`)
	return cmd
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
