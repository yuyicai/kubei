# kubei init

kubei init 分为三个阶段：runtime、kube、kubeadm
    runtime 是安装docker容器引擎
    kube 是安装k8s组件，包括kubeadm、kubelet、kubectl、kubernetes-cni、crictl
    kubeadm 是条用kubeadm对集群进行初始化，将nodes加入集群等工作，即创建集群这一步骤



## 分阶段部署

- 安装容器引擎

  ```
  kubei init phase runtime
  ```

- 安装安装k8s组件

  ```
  kubei init phase kube
  ```

- 创建集群

  ```
  kubei init phase kubeadm
  ```

