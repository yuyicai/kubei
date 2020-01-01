# ssh用户参数

```
--jump-server stringToString        Jump server user info, apply with --jump-server "host=IP,port=22,user=your-user,password=your-password,key=key-path" (default [])
    堡垒机配置，如果你执行kubei的机器ssh连接到需要部署集群的机器需要通过堡垒机，那么需要这个配置
    password和key两种认证方式二选一，实际使用时选择一个即可，如果两个都填写，将忽略password，只使用key
    配置示例：--jump-serve "--jump-server host=192.168.10.10,port=22,user=test,password=123456,key=$HOME/.ssh/jump.key"
    
--key string                        SSH key of the nodes.
    ssh连接s集群服务器的key，如果同时使用password和key，将忽略password，只使用key
    配置示例：key $HOME/.ssh/node.key

--password string                   SSH password of the nodes.
    ssh连接集群服务器的密码，如果同时使用password和key，将忽略password，只使用key
    如果使用普通用户部署，必须提供密码，因为sudo操作需要密码（如果你们普通用户sudo是免密的可省略）
    配置示例：--password 123456

--port string                       SSH port of the nodes. (default "22")
    ssh连接集群服务器的端口
    默认：22

--user string                       SSH user of the nodes. (default "root")
    ssh连接集群服务器的用户，如果是普通用户，那么该用户必须拥有sudo权限，并且使用--password参数提供sudo密码
    默认：root

```



# kubei init 参数

```
--masters strings                   The master nodes IP
    master节点 ip地址，可填写多个，使用英文的逗号隔开
    配置示例：--masters 10.3.0.10,10.3.0.11,10.3.0.12
    
--workers strings                   The worker nodes IP
    工作节点（即真正跑业务容器的节点） ip地址，可填写多个，使用英文的逗号隔开
    配置示例：--workers 10.3.0.20,10.3.0.21

--container-engine-version string   The Docker version.
    docker容器引擎版本，不加参数时使用最新版，版本支持18.09+
    配置示例：--container-engine-version 18.09.9
    
--kubernetes-version string         The Kubernetes version
    部署k8s集群所使用的kubernetes版本，执行1.16+
    配置示例：--kubernetes-version 1.16.4

--control-plane-endpoint string     Specify a DNS name for the control plane. (default "apiserver.k8s.local:6443")
    外部访问apiserver的地址，默认：apiserver.k8s.local:6443
    apiserver.k8s.local会被写到/etc/hosts,解析到127.0.0.1
    一般不需要更改这个地址

--image-repository string           Choose a container registry to pull control plane images from (default "gcr.azk8s.cn/google_containers")
    集群相关容器镜像仓库地址，从这个地址拉去的容器包括
    默认：gcr.azk8s.cn/google_containers

--pod-network-cidr string           Specify range of IP addresses for the pod network. If set, the control plane will automatically allocate CIDRs for every node. (default "10.244.0.0/16")
    k8s集群中pod的ip地址范围，一般不用更改

--service-cidr string               Use alternative range of IP address for service VIPs. (default "10.96.0.0/12")
    k8s集群中service地址范围，一般不用更改

--skip-phases strings               List of phases to be skipped
    跳过init中的某个步骤，这个与kubeadm中的用法一样
    init中包含了三个步骤（runtime、kube、kubeadm），使用使用"kubei init phase"进行查看
    runtime是部署docker容器引擎
    kube是部署k8s组件，包括kubeadm、kubelet、kubectl、kubernetes-cni、crictl
    kubeadm是条用kubeadm对集群进行初始化，将nodes加入集群等工作，即创建集群这一步骤


```



# kubei reset参数

```
--remove-container-engine         If true, remove the container engine from the nodes
    增加该参数将会删除容器引擎，后面不需要跟任何值，直接 --remove-container-engine 即可

--remove-kubernetes-component     If true, remove the kubernetes component from the nodes
    增加该参数将会kubernetes相关组件，后面不需要跟任何值，直接 --remove-kubernetes-component 即可
```

