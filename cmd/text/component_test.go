package text

import (
	"github.com/lithammer/dedent"
	"github.com/yuyicai/kubei/config/constants"
	"github.com/yuyicai/kubei/config/rundata"
	"testing"
)

func TestApt_Docker(t *testing.T) {
	type args struct {
		d rundata.Docker
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "(apt) docker install cmd",
			args: args{d: rundata.Docker{
				Version:        "18.09.9",
				CGroupDriver:   constants.DefaultCGroupDriver,
				LogDriver:      constants.DefaultLogDriver,
				LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
				StorageDriver:  constants.DockerDefaultStorageDriver,
			}},
			want: dedent.Dedent(`
				apt-get update -qq >/dev/null && DEBIAN_FRONTEND=noninteractive apt-get -y install apt-transport-https ca-certificates curl
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
				`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ap := Apt{}
			got, err := ap.Docker(tt.args.d)
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
		d rundata.Docker
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "(yum) docker install cmd",
			args: args{d: rundata.Docker{
				Version:        "18.09.9",
				CGroupDriver:   constants.DefaultCGroupDriver,
				LogDriver:      constants.DefaultLogDriver,
				LogOptsMaxSize: constants.DefaultLogOptsMaxSize,
				StorageDriver:  constants.DockerDefaultStorageDriver,
			}},
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
				      "https://dockerhub.azk8s.cn",
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yu := Yum{}
			got, err := yu.Docker(tt.args.d)
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
