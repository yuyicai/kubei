package image

import (
	"flag"
	"strings"
	"testing"

	"k8s.io/klog"
)

func TestDownloadFile(t *testing.T) {
	klog.InitFlags(nil)
	flag.Set("v", "8")
	flag.Parse()
	type args struct {
		imageUrl  string
		savePath  string
		cachePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "download from aliyun",
			args: args{
				imageUrl:  "registry.cn-hangzhou.aliyuncs.com/kubebin/kube-files:v1.20.0-test",
				savePath:  "/app/.kubei/kube/v1.20.0",
				cachePath: "/app/.kubei/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DownloadFile(tt.args.imageUrl, tt.args.savePath, tt.args.cachePath); (err != nil) != tt.wantErr {
				t.Errorf("DownloadFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDownloadImage(t *testing.T) {
	klog.InitFlags(nil)
	flag.Set("v", "8")
	flag.Parse()
	type args struct {
		imageUrl  string
		savePath  string
		cachePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "downlaod nginx 1.17.10",
			args: args{
				imageUrl:  "registry.cn-hangzhou.aliyuncs.com/kubebin/nginx:1.17.10",
				savePath:  "/app/.kubei/images",
				cachePath: "/app/.kubei/tmp",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DownloadImage(tt.args.imageUrl, tt.args.savePath, tt.args.cachePath); (err != nil) != tt.wantErr {
				t.Errorf("DownloadImage() error = %+v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestString(t *testing.T) {
	strImage := "test"
	stri := strings.Trim(strImage, "/")
	t.Log(stri)
	t.Run("test", func(t *testing.T) {
		ss := strings.SplitN(stri, "/", 2)
		t.Log(ss)
		t.Log(ss[len(ss)-1])
	})
}
