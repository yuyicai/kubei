package registry

import (
	"testing"
)

func TestDownloadFile(t *testing.T) {
	type args struct {
		imageUrl string
		user     string
		password string
		destPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "download from aliyun",
			args: args{
				imageUrl: "registry.cn-hangzhou.aliyuncs.com/kubebin/kube-files:v1.20.0-test",
				destPath: "/app/kube/v1.20.0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DownloadFile(tt.args.imageUrl, tt.args.user, tt.args.password, tt.args.destPath); (err != nil) != tt.wantErr {
				t.Errorf("DownloadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
