# 部署示例

## 使用密码作为ssh登录认证部署

```
./kubei init
 -p 123456 \
 -m 10.3.0.10,10.3.0.11,10.3.0.12 \
 -n 10.3.0.20,10.3.0.21 \
 -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
```

## 使用堡垒机

```
./kubei init \
 -k $HOME/.ssh/k8s.key \
 --jump-server "host=47.113.102.111,port=22,user=deer,key=$HOME/.ssh/jump.key" \
 -m 10.3.0.10,10.3.0.11,10.3.0.12 \
 -n 10.3.0.20,10.3.0.21 \
 -f ./kube_v1.17.9-docker_v18.09.9-flannel_v0.11.0-amd64.tgz
```


## 指定版本

在线安装可以选择版本，离线安装以离线包里的版本为准  
```
./kubei init \
 -k $HOME/.ssh/k8s.key \
 -m 10.3.0.10,10.3.0.11,10.3.0.12 \
 -n 10.3.0.20,10.3.0.21 \
 --kubernetes-version 1.16.4 \
 --container-engine-version 18.09.9
```


## 重置集群

```
./kubei reset \
 -k $HOME/.ssh/k8s.key \
 -m 10.3.0.10,10.3.0.11,10.3.0.12 \
 -n 10.3.0.20,10.3.0.21
```

## 重置所有

重置集群同时删除容器引擎和kubernetes相关组件

```
./kubei reset \
 -k $HOME/.ssh/k8s.key \
 -m 10.3.0.10,10.3.0.11,10.3.0.12 \
 -n 10.3.0.20,10.3.0.21 \
 --remove-container-engine \
 --remove-kubernetes-component
```
