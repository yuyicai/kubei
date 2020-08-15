package tmpl

import (
	"github.com/lithammer/dedent"
	"github.com/yuyicai/kubei/internal/constants"
	"github.com/yuyicai/kubei/internal/rundata"
	"testing"
)

func TestApt_Docker(t *testing.T) {
	type args struct {
		i string
		d rundata.Docker
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "(apt_docker) online install cmd",
			args: args{
				i: constants.InstallTypeOnline,
				d: rundata.Docker{
					Version:        "18.09.9",
					CGroupDriver:   constants.DefaultCGroupDriver,
					LogDriver:      constants.DefaultLogDriver,
					LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
					StorageDriver:  constants.DockerDefaultStorageDriver,
				},
			},
			want: dedent.Dedent(`
				apt-get update -qq >/dev/null && DEBIAN_FRONTEND=noninteractive apt-get -y install -qq apt-transport-https ca-certificates curl
				curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | apt-key add -qq - >/dev/null
				cat <<EOF | tee /etc/apt/sources.list.d/docker.list
				deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable
				EOF
				apt-get update -qq >/dev/null
				DOCKER_VER=$(apt-cache madison docker-ce | awk '/18.09.9/ {print$3}' | head -1)
				apt-get -y install -qq docker-ce=$DOCKER_VER docker-ce-cli=$DOCKER_VER containerd.io
				cat <<EOF | tee /etc/docker/daemon.json
				{
				  "registry-mirrors": [
				      "https://dockerhub.mirrors.nwafu.edu.cn/",
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
				`),
		},
		{
			name: "(apt_docker) online, not set version install cmd",
			args: args{
				i: constants.InstallTypeOnline,
				d: rundata.Docker{
					Version:        "",
					CGroupDriver:   constants.DefaultCGroupDriver,
					LogDriver:      constants.DefaultLogDriver,
					LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
					StorageDriver:  constants.DockerDefaultStorageDriver,
				},
			},
			want: dedent.Dedent(`
				apt-get update -qq >/dev/null && DEBIAN_FRONTEND=noninteractive apt-get -y install -qq apt-transport-https ca-certificates curl
				curl -fsSL https://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | apt-key add -qq - >/dev/null
				cat <<EOF | tee /etc/apt/sources.list.d/docker.list
				deb [arch=amd64] https://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable
				EOF
				apt-get update -qq >/dev/null
				apt-get -y install -qq docker-ce docker-ce-cli containerd.io
				cat <<EOF | tee /etc/docker/daemon.json
				{
				  "registry-mirrors": [
				      "https://dockerhub.mirrors.nwafu.edu.cn/",
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
				`),
		},
		{
			name: "(apt_docker) offline install cmd",
			args: args{
				i: constants.InstallTypeOffline,
				d: rundata.Docker{
					Version:        "18.09.9",
					CGroupDriver:   constants.DefaultCGroupDriver,
					LogDriver:      constants.DefaultLogDriver,
					LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
					StorageDriver:  constants.DockerDefaultStorageDriver,
				},
			},
			want: dedent.Dedent(`
				cat <<EOF | tee /etc/docker/daemon.json
				{
				  "registry-mirrors": [
				      "https://dockerhub.mirrors.nwafu.edu.cn/",
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
				sh /tmp/.kubei/container_engine/default.sh
				`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := Apt{}
			got, err := ap.Docker(tt.args.i, tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("Docker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Docker() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYum_Docker(t *testing.T) {
	type args struct {
		i string
		d rundata.Docker
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "(yum_docker) online install cmd",
			args: args{
				i: "online",
				d: rundata.Docker{
					Version:        "18.09.9",
					CGroupDriver:   constants.DefaultCGroupDriver,
					LogDriver:      constants.DefaultLogDriver,
					LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
					StorageDriver:  constants.DockerDefaultStorageDriver,
				},
			},
			want: dedent.Dedent(`
				yum install -y -q yum-utils
				yum-config-manager --add-repo \
				  https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
				DOCKER_VER=$(yum list docker-ce --showduplicates | awk '/18.09.9/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
				yum install -y -q docker-ce-$DOCKER_VER docker-ce-cli-$DOCKER_VER containerd.io
				mkdir -p /etc/docker
				cat <<EOF | tee /etc/docker/daemon.json
				{
				  "registry-mirrors": [
				      "https://dockerhub.mirrors.nwafu.edu.cn/",
				      "https://hub-mirror.c.163.com"
				  ],
				  "exec-opts": ["native.cgroupdriver=systemd"],
				  "log-driver": "json-file",
				  "log-opts": {
				    "max-size": "500m"
				  },
				  "storage-driver": "overlay2",
				  "storage-opts": [
				    "overlay2.override_kernel_check=true"
				  ]
				}
				EOF
				mkdir -p /etc/systemd/system/docker.service.d
				`),
		},
		{
			name: "(yum_docker) online, not set version install cmd",
			args: args{
				i: "online",
				d: rundata.Docker{
					Version:        "",
					CGroupDriver:   constants.DefaultCGroupDriver,
					LogDriver:      constants.DefaultLogDriver,
					LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
					StorageDriver:  constants.DockerDefaultStorageDriver,
				},
			},
			want: dedent.Dedent(`
				yum install -y -q yum-utils
				yum-config-manager --add-repo \
				  https://mirrors.aliyun.com/docker-ce/linux/centos/docker-ce.repo
				yum install -y -q docker-ce docker-ce-cli containerd.io
				mkdir -p /etc/docker
				cat <<EOF | tee /etc/docker/daemon.json
				{
				  "registry-mirrors": [
				      "https://dockerhub.mirrors.nwafu.edu.cn/",
				      "https://hub-mirror.c.163.com"
				  ],
				  "exec-opts": ["native.cgroupdriver=systemd"],
				  "log-driver": "json-file",
				  "log-opts": {
				    "max-size": "500m"
				  },
				  "storage-driver": "overlay2",
				  "storage-opts": [
				    "overlay2.override_kernel_check=true"
				  ]
				}
				EOF
				mkdir -p /etc/systemd/system/docker.service.d
				`),
		},
		{
			name: "(yum_docker) offline install cmd",
			args: args{
				i: "offline",
				d: rundata.Docker{
					Version:        "18.09.9",
					CGroupDriver:   constants.DefaultCGroupDriver,
					LogDriver:      constants.DefaultLogDriver,
					LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
					StorageDriver:  constants.DockerDefaultStorageDriver,
				},
			},
			want: dedent.Dedent(`
				mkdir -p /etc/docker
				cat <<EOF | tee /etc/docker/daemon.json
				{
				  "registry-mirrors": [
				      "https://dockerhub.mirrors.nwafu.edu.cn/",
				      "https://hub-mirror.c.163.com"
				  ],
				  "exec-opts": ["native.cgroupdriver=systemd"],
				  "log-driver": "json-file",
				  "log-opts": {
				    "max-size": "500m"
				  },
				  "storage-driver": "overlay2",
				  "storage-opts": [
				    "overlay2.override_kernel_check=true"
				  ]
				}
				EOF
				mkdir -p /etc/systemd/system/docker.service.d
				sh /tmp/.kubei/container_engine/default.sh
				`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yu := Yum{}
			got, err := yu.Docker(tt.args.i, tt.args.d)
			if (err != nil) != tt.wantErr {
				t.Errorf("Docker() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Docker() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApt_KubeComponent(t *testing.T) {
	type args struct {
		version     string
		installType string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "(apt_kubernetes) online install cmd",
			args: args{
				version:     "1.17.4",
				installType: constants.InstallTypeOnline,
			},
			want: dedent.Dedent(`
				apt-get update -qq && apt-get install -qq -y apt-transport-https curl
				curl -s https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add - >/dev/null
				cat <<EOF | tee /etc/apt/sources.list.d/kubernetes.list
				deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
				EOF
				apt-get update -qq
				KUBE_VER=$(apt-cache madison kubelet | awk '/1.17.4/ {print$3}' | head -1)
				apt-get install -qq -y --allow-change-held-packages kubelet=$KUBE_VER kubeadm=$KUBE_VER kubectl=$KUBE_VER
				apt-mark hold kubelet kubeadm kubectl
			`),
		},
		{
			name: "(apt_kubernetes) online, not set version install cmd",
			args: args{
				version:     "",
				installType: constants.InstallTypeOnline,
			},
			want: dedent.Dedent(`
				apt-get update -qq && apt-get install -qq -y apt-transport-https curl
				curl -s https://mirrors.aliyun.com/kubernetes/apt/doc/apt-key.gpg | apt-key add - >/dev/null
				cat <<EOF | tee /etc/apt/sources.list.d/kubernetes.list
				deb https://mirrors.aliyun.com/kubernetes/apt/ kubernetes-xenial main
				EOF
				apt-get update -qq
				apt-get install -qq -y --allow-change-held-packages kubelet kubeadm kubectl
				apt-mark hold kubelet kubeadm kubectl
			`),
		},
		{
			name: "(apt_kubernetes) offline install cmd",
			args: args{
				version:     "",
				installType: constants.InstallTypeOffline,
			},
			want: dedent.Dedent(`
				sh /tmp/.kubei/kubernetes/default.sh
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := Apt{}
			got, err := ap.KubeComponent(tt.args.version, tt.args.installType)
			if (err != nil) != tt.wantErr {
				t.Errorf("KubeComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KubeComponent() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYum_KubeComponent(t *testing.T) {
	type args struct {
		version     string
		installType string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "(yum_kubernetes) online install cmd",
			args: args{
				version:     "1.17.4",
				installType: constants.InstallTypeOnline,
			},
			want: dedent.Dedent(`
				cat <<EOF | tee /etc/yum.repos.d/kubernetes.repo
				[kubernetes]
				name=Kubernetes
				baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
				enabled=1
				gpgcheck=1
				repo_gpgcheck=1
				gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
				EOF
				setenforce 0 || true
				sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
				KUBE_VER=$(yum list kubelet --showduplicates | awk '/1.17.4/ {print$2}' | tail -1 | sed 's/[[:digit:]]://')
				yum install -y kubelet-$KUBE_VER kubeadm-$KUBE_VER kubectl-$KUBE_VER --disableexcludes=kubernetes
			`),
		},
		{
			name: "(yum_kubernetes) online, not set version install cmd",
			args: args{
				version:     "",
				installType: constants.InstallTypeOnline,
			},
			want: dedent.Dedent(`
				cat <<EOF | tee /etc/yum.repos.d/kubernetes.repo
				[kubernetes]
				name=Kubernetes
				baseurl=https://mirrors.aliyun.com/kubernetes/yum/repos/kubernetes-el7-x86_64
				enabled=1
				gpgcheck=1
				repo_gpgcheck=1
				gpgkey=https://mirrors.aliyun.com/kubernetes/yum/doc/yum-key.gpg https://mirrors.aliyun.com/kubernetes/yum/doc/rpm-package-key.gpg
				EOF
				setenforce 0 || true
				sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
				yum install -y kubelet kubeadm kubectl --disableexcludes=kubernetes
			`),
		},
		{
			name: "(yum_kubernetes) offline install cmd",
			args: args{
				version:     "",
				installType: constants.InstallTypeOffline,
			},
			want: dedent.Dedent(`
				setenforce 0 || true
				sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config
				sh /tmp/.kubei/kubernetes/default.sh
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yu := Yum{}
			got, err := yu.KubeComponent(tt.args.version, tt.args.installType)
			if (err != nil) != tt.wantErr {
				t.Errorf("KubeComponent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("KubeComponent() got = %v, want %v", got, tt.want)
			}
		})
	}
}
