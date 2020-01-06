# kubei

`kubei` （kubernetes installer） 是一个go开发的用来部署kubernetes高可用集群的命令行工具，该工具可在`Windows`、`Linux`、`Mac`中运行  

`kubei`原理：通过ssh连接到集群服务器，进行容器引擎安装、kubernetes组件安装、调用kubeadm初始化集群  

支持使用普通用户（sudo用户）连接集群服务器进行安装部署，支持通过堡垒机连接集群服务器  



# 版本支持

<table>
    <thead>
        <tr>
            <th align="center" colspan="2">应用</th>
            <th align="center">版本</thalign="center">
        </tr>
    </thead>
    <tbody>
        <tr>
            <td align="center" colspan="2">kubernetes</td>
            <td align="center">1.16.X、1.17.X</td>
        </tr>
        <tr>
            <td align="center">容器引擎</td>
            <td align="center">目前只支持docker</td>
            <td align="center">18.09.X、19.XX.XX</td>
        </tr>
        <tr>
            <td align="center">网络插件</td>
            <td align="center">目前只支持flannel</td>
            <td align="center">0.11.0</td>
        </tr>
    </tbody>
</table>




# 快速开始

|   主机    | 集群角色 |      系统版本      |
| :-------: | :------: | :----------------: |
| 10.3.0.10 |  master  | Ubuntu 18.04.3 LTS |
| 10.3.0.11 |  master  | Ubuntu 16.04.6 LTS |
| 10.3.0.12 |  master  |     CentOS 7.4     |
| 10.3.0.20 |  worker  |     CentOS 7.7     |
| 10.3.0.21 |  worker  | Ubuntu 18.04.3 LTS |

*默认使用root用户和22端口，如果需要使用普通用户和其它ssh端口，请查看[ssh用户参数说明](./docs/flags.md)*

*如果要用密码做ssh登录验证，请查看[ssh用户参数说明](./docs/flags.md)*

**执行部署命令：**

```
kubei init --key=$HOME/.ssh/k8s.key \
 --masters 10.3.0.10,10.3.0.11,10.3.0.12 \
 --workers 10.3.0.20,10.3.0.21 \
 --skip-headers
```

部署过程：

[![asciicast](https://asciinema.org/a/291242.svg)](https://asciinema.org/a/291242)

部署结果：

```
NAME        STATUS   ROLES    AGE   VERSION   INTERNAL-IP   EXTERNAL-IP   OS-IMAGE                KERNEL-VERSION               CONTAINER-RUNTIME
10.3.0.10   Ready    master   58s   v1.17.0   10.3.0.10     <none>        Ubuntu 18.04.3 LTS      4.15.0-66-generic            docker://19.3.5
10.3.0.11   Ready    master   21s   v1.17.0   10.3.0.11     <none>        Ubuntu 16.04.6 LTS      4.4.0-142-generic            docker://19.3.5
10.3.0.12   Ready    master   28s   v1.17.0   10.3.0.12     <none>        CentOS Linux 7 (Core)   3.10.0-1062.1.2.el7.x86_64   docker://19.3.5
10.3.0.20   Ready    <none>   34s   v1.17.0   10.3.0.20     <none>        CentOS Linux 7 (Core)   3.10.0-693.2.2.el7.x86_64    docker://19.3.5
10.3.0.21   Ready    <none>   11s   v1.17.0   10.3.0.21     <none>        Ubuntu 18.04.3 LTS      4.15.0-66-generic            docker://19.3.5
```



[更多安装示例](./docs/example.md)（指定安装版本，使用堡垒机连接等）

[参数说明](./docs/flags.md)



TODO

- [ ] calico网络组件支持
- [ ] 增加节点功能
- [ ] 离线部署