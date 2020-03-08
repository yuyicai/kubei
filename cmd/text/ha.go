package text

import (
	"bytes"
	"fmt"
	"github.com/lithammer/dedent"
	"text/template"
)

func KubeletUnitFile(image string) string {
	cmdText := dedent.Dedent(`
        cgroupDriver=$(docker info --format '{{json .CgroupDriver}}' | sed 's/"//g')
        mkdir -p /etc/systemd/system/kubelet.service.d
        cat << EOF | tee /etc/systemd/system/kubelet.service.d/20-ha-service-manager.conf
        [Service]
        ExecStart=
        ExecStart=/usr/bin/kubelet --address=127.0.0.1 --pod-manifest-path=/etc/kubernetes/manifests --pod-infra-container-image=%s --cgroup-driver=${cgroupDriver}
        Restart=always
        EOF
	`)

	return fmt.Sprintf(cmdText, image)
}

func RemoveKubeletUnitFile() string {
	cmdText := "rm -f /etc/systemd/system/kubelet.service.d/20-ha-service-manager.conf"
	return cmdText
}

func NginxConf(masters []string, nginxPort, masterPort string) (string, error) {
	m := map[string]interface{}{
		"masters":    masters,
		"nginxPort":  nginxPort,
		"masterPort": masterPort,
	}

	cmdText := dedent.Dedent(`
        mkdir -p /etc/kubernetes
        cat <<EOF | tee /etc/kubernetes/nginx.conf
        error_log stderr notice;
        
        worker_processes 2;
        worker_rlimit_nofile 130048;
        worker_shutdown_timeout 10s;
        
        events {
          multi_accept on;
          use epoll;
          worker_connections 16384;
        }
        
        stream {
          upstream kube_apiserver {
            least_conn;
        {{range $master := .masters}}
            server {{ $master }}:{{ .masterPort }};
        {{- end}}
          }
        
          server {
            listen        127.0.0.1:{{ .nginxPort }};
            proxy_pass    kube_apiserver;
            proxy_timeout 10m;
            proxy_connect_timeout 1s;
          }
        }
        
        http {
          aio threads;
          aio_write on;
          tcp_nopush on;
          tcp_nodelay on;
        
          keepalive_timeout 5m;
          keepalive_requests 100;
          reset_timedout_connection on;
          server_tokens off;
          autoindex off;
        }
        EOF
	`)

	t, err := template.New("text").Parse(cmdText)
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

func NginxManifest(nginxImage string) string {
	cmdText := dedent.Dedent(`
        mkdir -p /etc/kubernetes/manifests
        cat <<EOF | tee /etc/kubernetes/manifests/nginx-proxy.yml
        apiVersion: v1
        kind: Pod
        metadata:
          name: nginx-proxy
          namespace: kube-system
          labels:
            addonmanager.kubernetes.io/mode: Reconcile
            k8s-app: kube-nginx
        spec:
          hostNetwork: true
          dnsPolicy: ClusterFirstWithHostNet
          nodeSelector:
            beta.kubernetes.io/os: linux
          priorityClassName: system-node-critical
          containers:
          - name: nginx-proxy
            image: %s
            imagePullPolicy: IfNotPresent
            resources:
              requests:
                cpu: 25m
                memory: 32M
            securityContext:
              privileged: true
            volumeMounts:
            - mountPath: /etc/nginx/nginx.conf
              name: nginx-conf
              readOnly: true
          volumes:
          - name: nginx-conf
            hostPath:
              path: /etc/kubernetes/nginx.conf
              type: FileOrCreate
        EOF
	`)
	return fmt.Sprintf(cmdText, nginxImage)
}
