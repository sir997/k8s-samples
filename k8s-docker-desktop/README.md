### k8s v1.25.2

### Images
> 问题：coredns无法pull，其它image无问题，导致卡在starting...,解决方案:
> `docker pull coredns:1.9.3 && docker tag coredns:1.9.3 k8s.gcr.io/coredns:v1.9.3`
```
k8s.gcr.io/conformance:v1.25.2
k8s.gcr.io/kube-apiserver:v1.25.2
k8s.gcr.io/kube-controller-manager:v1.25.2
k8s.gcr.io/kube-proxy:v1.25.2
k8s.gcr.io/kube-scheduler:v1.25.2
k8s.gcr.io/pause:3.8
k8s.gcr.io/etcd:3.5.4-0
k8s.gcr.io/coredns:v1.9.3
```