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
  cpu: 1
  memory: 1G
```

```
$ kubectl apply -f config/samples/perf_v1_baseline.yaml
baseline.perf.baseline.io/baseline-sample configured

$ kubectl get baseline
NAME              AGE
baseline-sample   1m
```

Check for the daemonset:
```
$ oc kubectl daemonset
NAME              DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
baseline-sample   1         1         1       1            1           <none>          1m
```

Check for the pods:
```
$ kubectl get pods
NAME                    READY   STATUS    RESTARTS   AGE
baseline-sample-nnq5b   1/1     Running   0          1m

$ kubectl logs baseline-sample-nnq5b 
stress-ng: info:  [1] setting to a 0 second run per stressor
stress-ng: info:  [1] dispatching hogs: 1 cpu, 1 vm
```

The resulting command is stored in the status of the CRD:
```
$ kubectl get -o template baseline/baseline-sample --template={{.status.command}}
stress-ng --timeout 0 --cpu 1 --vm 1 --vm-bytes 1G
```

Update a parameter of the CRD:
```
$ kubectl patch baseline baseline-sample --type merge -p '{"spec":{"cpu":2}}'
baseline.perf.baseline.io/baseline-sample patched

$ kubectl get po
NAME                    READY   STATUS              RESTARTS   AGE
baseline-sample-nnq5b   1/1     Terminating         0          5m
baseline-sample-xvxc9   0/1     ContainerCreating   0          1s

$ kubectl logs baseline-sample-xvxc9
stress-ng: info:  [1] setting to a 0 second run per stressor
stress-ng: info:  [1] dispatching hogs: 2 cpu, 1 vm
```

## Installation

```
$ git clone https://github.com/josecastillolema/baseline-operator
$ cd baseline-operator
$ make deploy IMG="josecastillolema/baseline-operator:v0.6"
```
