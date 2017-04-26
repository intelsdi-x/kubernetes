### Cgroup-cpuset-cpus isolator.

This is an example implementation of extended pod-level and/or container-level resource isolation as specified [here](https://github.com/kubernetes/features/issues/246).
Cgroup-cpuset-cpus isolator aims to provide a way for an operator to specify which containers in the POD should be run with exclusive [cpusets](http://man7.org/linux/man-pages/man7/cpuset.7.html).

### Building 

In order to use this feature, one has to build a modified kubelet which resides in Intel's fork of Kubernetes [here](https://github.com/intelsdi-x/kubernetes/tree/ext-iso).

First issue:

```
hack/update-generated-runtime.sh
```

to generate lifecycle protobuf files, and then compile kubelet with:


```
build/run.sh make KUBE_FASTBUILD=true ARCH=amd64
```

a kubelet binary can be found here:

```
_output/dockerized/bin/linux/amd64/kubelet
```

Every node which is supposed to be able to use cgroup-cpuset-cpus isolator has to use this modified kubelet.
To enable extended pod-level resource isolation kubelet has to be run with two additional flags (true by default since 1.6):

```
--cgroups-per-qos=true --cgroup-root=/
```

Kubelet spawns a gRPC server in-process that implements a new interface called EventDispatcher. Isolators register themselves with this service to receive isolation events. It is convenient to run isolators as DaemonSet pods, orchestrating placement via the usual means (node labels and NodeSelectors, taints, tolerations, etc.)


To create a docker image with cgroup-cpuset-cpus isolator run (resulting docker images name is passed as IMAGE_NAME variable):

```sh
IMAGE_NAME=<docker_image_name> make docker
```

### Usage

Example manifests for cgroup-cpuset-cpus isolator can be found in cluster/addons/iso-client/examples. We also provide an RBAC configuration for isolator to be able to advertise its own capacity via opaque integer resources. In order to spawn it on desired nodes in kubernetes cluster, label your nodes where you intend to run with cgroup-cpuset-cpus-isolator=enabled label:

```
kubectl label node <node-name> cgroup-cpuset-cpus-isolator=enabled
```

Then create cgroup-cpuset-cpus-isolator and rbac configuration. modify the image section with previously built image( preferably available in private docker registry or all nodes which will use this isolator) and run:

```
kubectl create -f cluster/addons/iso-client/examples
```

When isolator is spawned it will advertise opaque integer resources onto the node where it is running. One can check them by issuing:

```
kubectl describe node <node-name>
```

In a Capacity section, there should be:

```
pod.alpha.kubernetes.io/opaque-int-resource-cgroup-cpuset-cpus: <number-of-cpus-for-cpusets>
```

To run a pod with specified cpusets, one just have to modify manifests of existing applications to consume those opaque integer resources, example:

```yaml
kind: Pod
apiVersion: v1
metadata:
  name: app-with-pinning
  labels:
    app: example
spec:
  containers:
  - name: app-with-pinning
    image: <my-app-image
    resources:
      requests:
        pod.alpha.kubernetes.io/opaque-int-resource-cgroup-cpuset-cpus: 3
```
