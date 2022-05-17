![GitHub](https://img.shields.io/github/license/josecastillolema/baseline-operator)
![GitHub language count](https://img.shields.io/github/languages/count/josecastillolema/baseline-operator)
![GitHub top language](https://img.shields.io/github/languages/top/josecastillolema/baseline-operator)
[![Docker Image CI](https://github.com/josecastillolema/baseline-operator/actions/workflows/docker-image.yml/badge.svg)](https://github.com/josecastillolema/baseline-operator/actions/workflows/docker-image.yml)

# baseline-operator

The baseline-operator provides a way to run [stress-ng](https://wiki.ubuntu.com/Kernel/Reference/stress-ng) workloads on Kubernetes in various deployment configurations.

It is intended to create an artificial baseline load in a Kubernetes cluster in order to be able to run other Kubernetes tests in more realistic conditions (performance, cluster upgrade, etc.).

## Use

Create a baseline CRD, i.e.:
```yaml
apiVersion: perf.baseline.io/v1
kind: Baseline
metadata:
  name: baseline-sample
spec:
  cpu: 1                                             # cores
  mem: 1G                                            # size of the virtual memory. can be defined as a % of the available memory
  io: 1                                              # workers continuously calling sync to commit buffer cache to disk
  sock: 1                                            # workers exercising socket I/O networking
  custom: "--timer 1"                                # other custom params
  # image: quay.io/cloud-bulldozer/stressng          # custom image
  # hostNetwork: true                                # directly use host network
  # nodeSelector:                                    # filter nodes with labels
  #   stress: "true"
  # tolerations:                                     # use the control plane nodes
  # - key: node-role.kubernetes.io/control-plane
  #   operator: Exists
  #   effect: NoSchedule
  # - key: node-role.kubernetes.io/master
  #   operator: Exists
  #   effect: NoSchedule
```

```
$ kubectl apply -f config/samples/perf_v1_baseline.yaml
baseline.perf.baseline.io/baseline-sample configured

$ kubectl get baseline
NAME              COMMAND                                                                                AGE
baseline-sample   stress-ng -t 0 --cpu 1 --vm 1 --vm-bytes 1G --io 1 --sock 1 --sock-if eth0 --timer 1   3s
```

Check for the DaemonSet:
```
$ kubectl get daemonset
NAME              DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
baseline-sample   1         1         1       1            1           <none>          1m
```

Check for the Pods:
```
$ kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
baseline-sample-nnq5b   1/1     Running   0          1m

$ kubectl logs baseline-sample-nnq5b 
stress-ng: info:  [1] setting to a 0 second run per stressor
stress-ng: info:  [1] dispatching hogs: 1 cpu, 1 vm, 1 io, 1 sock, 1 timer
```

### Updating the CRD

Update or remove parameters from the CRD:
```
$ kubectl patch baseline baseline-sample --type merge -p '{"spec":{"cpu":2}}'
baseline.perf.baseline.io/baseline-sample patched

$ kubectl get po
NAME                    READY   STATUS              RESTARTS   AGE
baseline-sample-nnq5b   1/1     Terminating         0          5m
baseline-sample-xvxc9   0/1     ContainerCreating   0          1s

$ kubectl logs baseline-sample-xvxc9
stress-ng: info:  [1] setting to a 0 second run per stressor
stress-ng: info:  [1] dispatching hogs: 2 cpu, 1 vm, 1 io, 1 sock, 1 timer
```

Some updates (like the above) that get translated into a new command cause the DaemonSet to be recreated.

Check for the CRD events:
```
$ kubectl describe baseline baseline-sample
...
Events:
  Type    Reason     Age    From      Message
  ----    ------     ----   ----      -------
  Normal  Created    5m20s  Baseline  Created daemonset default/baseline-sample
  Normal  Recreated  4s     Baseline  Rereated daemonset default/baseline-sample
```

Other fields of the CRD can be updated without the need of a DaemonSet recreation, like i.e.: `image`, `hostNetwork`, `nodeSelector` and `tolerations`:

```
$ kubectl patch baseline baseline-sample --type merge -p '{"spec":{"nodeSelector":{"stress":"true"}}}'
baseline.perf.baseline.io/baseline-sample patched

$ kubects get daemonset
NAME              DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
baseline-sample   0         0         0       0            0           stress=true     3m9s

$ kubectl describe baseline baseline-sample
...
Events:
  Type    Reason     Age    From      Message
  ----    ------     ----   ----      -------
  Normal  Created    6m20s  Baseline  Created daemonset default/baseline-sample
  Normal  Recreated  1m10s  Baseline  Rereated daemonset default/baseline-sample
  Normal  Updated    7s     Baseline  Updated daemonset default/baseline-sample
```

### Node placement

If you specify node selector(s), then the DaemonSet controller will create Pods on nodes which match that node selector(s):
```yaml
apiVersion: perf.baseline.io/v1
kind: Baseline
metadata:
  name: baseline-sample
spec:
  cpu: 1
  nodeSelector:
    stress: "true"
```

```
$ kubectl get daemonset
NAME              DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
baseline-sample   1         1         1       1            1           stress=true     1m
```

By default, DaemonSet Pods only run in worker nodes. If you want to run *stress-ng* loads in control plane nodes you can use tolerations:
```yaml
apiVersion: perf.baseline.io/v1
kind: Baseline
metadata:
  name: baseline-sample
spec:
  cpu: 1			            
  tolerations:
  # these tolerations are to have the daemonset runnable on control plane nodes
  # remove them if your control plane nodes should not run pods
  - key: node-role.kubernetes.io/control-plane
    operator: Exists
    effect: NoSchedule
  - key: node-role.kubernetes.io/master
    operator: Exists
    effect: NoSchedule
```

### Custom image

It is possible to select a custom image for *stress-ng* using the `image` property:
```yaml
apiVersion: perf.baseline.io/v1
kind: Baseline
metadata:
  name: baseline-sample
spec:
  cpu: 1
  image: quay.io/cloud-bulldozer/stressng
```

If not defined, defaults to `quay.io/jcastillolema/stressng:0.14.01`. For network workloads is important for *stress-ng* version to be >= 0.14.01, which allows to choose the network interface via `--sock-if`, `sockmany-if`, `--udp-if` and `udp-flood-if`. The default image was compiled through this Dockerfile:
```Dockerfile
FROM quay.io/centos/centos:stream8

WORKDIR /root
RUN yum install -y libaio-devel libattr-devel libcap-devel libgcrypt-devel libjpeg-devel keyutils-libs-devel lksctp-tools-devel libatomic zlib-devel cmake gcc
RUN curl -L https://github.com/ColinIanKing/stress-ng/archive/refs/tags/V0.14.01.tar.gz -o V0.14.01.tar.gz && tar -xzvf V0.14.01.tar.gz -C /root --strip-components=1
RUN make clean && make && mv stress-ng /usr/local/bin
```

## Installation

```
$ git clone https://github.com/josecastillolema/baseline-operator
$ cd baseline-operator
$ make deploy
```
