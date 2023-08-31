### 名词解释
- OCI(Open Container Initiaiv):容器化标准，[官网](https://opencontainers.org/)
> OCI主要包含两个规范，一个是容器运行时规范(runtime-spec)，一个是容器镜像规范(image-spec)。
-  CNI(Container Network Interface):容器网络接口，[官网](https://www.cni.dev/)，[规范](https://github.com/containernetworking/cni/blob/spec-v0.4.0/SPEC.md)，CNI规范文档主要用来说明容器运行时(runtimes)和插件(plugins)之间的接口。
> CNI只关心容器创建时的网络分配，以及当容器被删除时已经分配网络资源的释放。

### 重学容器系列
[Container](https://blog.frognew.com/2021/04/relearning-container-01.html)

### 网络

#### 跨主机通讯

- overlay
- vxlan
- bgp
- macvlan