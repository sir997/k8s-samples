## CNI

#### CNI插件安装
```
wget https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz

mkdir -p /opt/cni/bin/
mkdir -p /etc/cni/net.d

tar -zxvf cni-plugins-linux-amd64-v0.9.1.tgz -C /opt/cni/bin/

tree /opt/cni/bin/
/opt/cni/bin/
├── bandwidth
├── bridge
├── dhcp
├── firewall
├── flannel
├── host-device
├── host-local
├── ipvlan
├── loopback
├── macvlan
├── portmap
├── ptp
├── sbr
├── static
├── tuning
├── vlan
└── vrf
```

##### 这些可执行文件从功能的角度可以划分为以下三类:
- 主插件: 用于创建网络设备

    - bridge: 创建一个网桥设备，并添加宿主机和容器到该网桥
    - ipvlan: 为容器添加ipvlan网络接口
    - loopback: 设置lo网络接口的状态为up
    - macvlan: 创建一个新的MAC地址，并将所有流量转发到容器
    - ptp: 创建Veth对
    - vlan: 分配一个vlan设备
    - host-device: 将已存在的设备移入容器内

- IPAM插件: 用于IP地址的分配

    - dhcp: 在宿主机上运行dhcp守护程序，代表容器发出dhcp请求
    - host-local: 维护一个分配ip的本地数据库
    - static: 为容器分配一个静态IPv4/IPv6地址，主要用于调试

- Meta插件: 其他插件，非单独使用插件

    - flannel: flannel网络方案的CNI插件，根据flannel的配置文件创建网络接口
    - tuning: 调整现有网络接口的sysctl参数
    - portmap: 一个基于iptables的portmapping插件。将端口从主机的地址空间映射到容器
    - bandwidth: 允许使用TBF进行限流的插件
    - sbr: 一个为网络接口配置基于源路由的插件
    - firewall: 过iptables给容器网络的进出流量进行一系列限制的插件

cnitool
```
git clone https://github.com/containernetworking/cni.git
cd cni
go mod tidy
cd cnitool
GOOS=linux GOARCH=amd64 go build .

chmod +x /opt/cni/bin/cnitool
ln -s /opt/cni/bin/cnitool /usr/local/bin/cnitool

cnitool
cnitool: Add, check, or remove network interfaces from a network namespace
  cnitool add   <net> <netns>
  cnitool check <net> <netns>
  cnitool del   <net> <netns>
```

#### 创建容器网络
创建containerd容器使用cni的配置文件:
```
cat << EOF | tee /etc/cni/net.d/redisnet.conf
{
    "cniVersion": "0.4.0",
    "name": "redisnet",
    "type": "bridge",
    "bridge": "cni0",
    "isDefaultGateway": true,
    "forceAddress": false,
    "ipMasq": true,
    "hairpinMode": true,
    "ipam": {
        "type": "host-local",
        "subnet": "10.88.0.0/16"
    }
}
EOF
```
创建一个名为redisnet的network namespace：
```
ip netns add redisnet

ip netns list
redisnet

ls /var/run/netns/
redisnet
```

向这个network namespace中添加网络:
```
export CNI_PATH=/opt/cni/bin
cnitool add redisnet /var/run/netns/redisnet
cnitool check redisnet /var/run/netns/redisnet
```

测试网络是否工作:
```
ip -n redisnet addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN qlen 1
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
3: eth0@if7: <BROADCAST,MULTICAST,UP,LOWER_UP> mtu 1500 qdisc noqueue state UP
    link/ether be:73:a7:ae:18:7f brd ff:ff:ff:ff:ff:ff link-netnsid 0
    inet 10.88.0.2/16 brd 10.88.255.255 scope global eth0
       valid_lft forever preferred_lft forever
    inet6 fe80::bc73:a7ff:feae:187f/64 scope link
       valid_lft forever preferred_lft forever

ping 10.88.0.2
```

#### 启动带网络的容器
ctr run命令在启动容器的时候可以使用--with-ns选项让容器在启动时候加入到一个已经存在的一个linux namespace，这里加入的是起那么创建的redisnet这个网络namespace。
```
ctr run --with-ns=network:/var/run/netns/redisnet -d docker.io/library/redis:alpine3.13 redis
```

删除容器后，可以按照下面的步骤清理网络资源：
```
export CNI_PATH=/opt/cni/bin
cnitool del redisnet /var/run/netns/redisnet
ip netns del redisnet
```