# PodCount K8s Operator

This is a simple Kubernates placeholder Operator. Its purpose it to define a Custom Resource `PodCount`,
whose purpose is to define in its CRD a `podsPerNode` integer value which represents the maximum number of pods
of a certain type per node. The type of pod is defined using `name` string in the CRD.

At reconciliation time whether limit of pods per node and type is overtaken, a `maxPodsExceeded` boolean value
is turned to `true` (very naif at the moment).
