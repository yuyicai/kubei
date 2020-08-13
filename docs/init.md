# kubei init

kubei init 分为五个阶段：`send`、`container-engine`、`kube`、`kubeadm`  
- `send` 离线安装时分发离线包到各节点
- `container-engine` 安装docker容器引擎
- `kube` 安装k8s组件，包括kubeadm、kubelet、kubectl、kubernetes-cni、crictl
- `cert` 签发证书，以替代kubeadm签发的证书，可自定义证书过期时间
- `kubeadm` 调用kubeadm对集群进行初始化，将nodes加入集群



## 分阶段部署

- 分发离线安装包

  ```
  kubei init phase send \
    -k $HOME/.ssh/k8s.key \
    -m 10.3.0.10,10.3.0.11,10.3.0.12 \
    -n 10.3.0.20,10.3.0.21 \
    -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
  ```

- 安装容器引擎

  ```
  kubei init phase container-engine \
    -k $HOME/.ssh/k8s.key \
    -m 10.3.0.10,10.3.0.11,10.3.0.12 \
    -n 10.3.0.20,10.3.0.21 \
    -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
  ```

- 安装安装k8s组件

  ```
  kubei init phase kube \
    -k $HOME/.ssh/k8s.key \
    -m 10.3.0.10,10.3.0.11,10.3.0.12 \
    -n 10.3.0.20,10.3.0.21 \
    -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
  ```

- 签发证书

  ```
  kubei init phase cert \
    -k $HOME/.ssh/k8s.key \
    -m 10.3.0.10,10.3.0.11,10.3.0.12 \
    -n 10.3.0.20,10.3.0.21 \
    -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
  ```

- 创建集群

  ```
  kubei init phase kubeadm \
    -k $HOME/.ssh/k8s.key \
    -m 10.3.0.10,10.3.0.11,10.3.0.12 \
    -n 10.3.0.20,10.3.0.21 \
    -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
  ```

