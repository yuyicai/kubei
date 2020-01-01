# 部署示例

*因为是示例，所以多用几个系统版本，实际部署中，最好还是统一系统版本*

|   主机    | 集群角色 |      系统版本      |
| :-------: | :------: | :----------------: |
| 10.3.0.10 |  master  | Ubuntu 18.04.3 LTS |
| 10.3.0.11 |  master  | Ubuntu 16.04.6 LTS |
| 10.3.0.12 |  master  |     CentOS 7.4     |
| 10.3.0.20 |  worker  |     CentOS 7.7     |
| 10.3.0.21 |  worker  | Ubuntu 18.04.3 LTS |



## 使用key作为ssh登录认证部署

```
./kubei init --key=$HOME/.ssh/k8s.key \
 --masters 10.3.0.10,10.3.0.11,10.3.0.12 \
 --workers 10.3.0.20,10.3.0.21 \
 --skip-headers
```

[![asciicast](https://asciinema.org/a/291242.svg)](https://asciinema.org/a/291242)



## 使用堡垒机

```
./kubei init --key=$HOME/.ssh/k8s.key \
 --jump-server "host=47.113.102.111,port=22,user=deer,key=$HOME/.ssh/jump.key" \
 --masters 10.3.0.10,10.3.0.11,10.3.0.12 \
 --workers 10.3.0.20,10.3.0.21 \
 --skip-headers
```

[![asciicast](https://asciinema.org/a/291262.svg)](https://asciinema.org/a/291262)



## 指定版本

```
./kubei init --key=$HOME/.ssh/k8s.key \
 --masters 10.3.0.10,10.3.0.11,10.3.0.12 \
 --workers 10.3.0.20,10.3.0.21 \
 --kubernetes-version 1.16.4 \
 --container-engine-version 18.09.9 \
 --skip-headers
```

[![asciicast](https://asciinema.org/a/291263.svg)](https://asciinema.org/a/291263)



## 重置集群

```
./kubei reset --key=$HOME/.ssh/k8s.key \
 --masters 10.3.0.10,10.3.0.11,10.3.0.12 \
 --workers 10.3.0.20,10.3.0.21 \
 --skip-headers
```

[![asciicast](https://asciinema.org/a/291265.svg)](https://asciinema.org/a/291265)



## 重置所有

重置集群同时删除容器引擎和kubernetes相关组件

```
./kubei reset --key=$HOME/.ssh/k8s.key \
 --masters 10.3.0.10,10.3.0.11,10.3.0.12 \
 --workers 10.3.0.20,10.3.0.21 \
 --remove-container-engine \
 --remove-kubernetes-component \
 --skip-headers
```

[![asciicast](https://asciinema.org/a/291266.svg)](https://asciinema.org/a/291266)

