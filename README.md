# PodCount K8s Operator

This is a simple Kubernates placeholder Operator base on the `kubebuilder` framework.
Its purpose is to define a Custom Resource `PodCount`, whose purpose is to describe in its CRD a `podsPerNode` integer value.
The latter represents the maximum number of pods of a certain type per node. The type of pod is defined using `name` string in the CRD.

At reconciliation time whether limit of pods per node and type is overtaken, a `maxPodsExceeded` boolean value
is turned to `true` (very naif at the moment).

## Prerequisites

* go version v1.20.0+
* docker version 17.03+.
* kubectl version v1.11.3+.
* Access to a Kubernetes v1.11.3+ cluster.

## Installation

Install the CRD into the cluster.

```sh
make install
```

## Run

Start Controller operating reconciliation.

```sh
make run
```

## Test

```sh
make test
```

### Other operations

* Deploy the Manager to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/k8s-hello-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself
cluster-admin privileges or be logged in as admin.

* Create instances of your solution*
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

### To Uninstall

* Delete the instances (CRs) from the cluster:

```sh
kubectl delete -k config/samples/
```

* Delete the APIs(CRDs) from the cluster:

```sh
make uninstall
```

* UnDeploy the controller from the cluster:

```sh
make undeploy
```

## See Also

* [Introduction - The Kubebuilder Book](https://book.kubebuilder.io/introduction)

## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

<http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
