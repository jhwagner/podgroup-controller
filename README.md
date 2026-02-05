# podgroup-controller

A Kubernetes controller that manages pod groups, waiting for all pods in a group to reach a specified condition before taking action. Designed for simulating gang scheduling with KWOK.

## Description

This controller watches Pods labeled with a pod group identifier and applies a "ready" label once all pods in the group reach the Running state. This is useful for:

- **Gang Scheduling Simulations**: Combined with KWOK, simulate ML training workloads where all pods must be running before work begins
- **Multi-pod Coordination**: Coordinate actions across related pods (Jobs, JobSets, StatefulSets, etc.)
- **Deadlock Demonstrations**: Show what happens when partial scheduling blocks cluster resources

### How It Works

1. Label your pods with `podgroup.jhwagner.github.io/name=<group-name>`
2. The controller watches for pods with this label
3. When all pods in a group are Running, it applies `podgroup.jhwagner.github.io/ready=true` to each pod
4. (optional) KWOK stages can then select on this label to transition pods to completion

### Example

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: worker-1
  labels:
    podgroup.jhwagner.github.io/name: training-job-1
spec:
  containers:
  - name: worker
    image: fake-image
```

Once all pods with `podgroup.jhwagner.github.io/name: training-job-1` are Running, the controller adds:
```yaml
labels:
  podgroup.jhwagner.github.io/ready: "true"
```

## Installation

### Quick Install

Deploy the latest version to your cluster:

```sh
kubectl apply -f https://raw.githubusercontent.com/jhwagner/podgroup-controller/main/dist/install.yaml
```

### Uninstall

```sh
kubectl delete -f https://raw.githubusercontent.com/jhwagner/podgroup-controller/main/dist/install.yaml
```

## Development

### Prerequisites
- go version v1.23+
- kubectl and access to a Kubernetes cluster
- docker (for building images)

### Run Locally

For development and testing with your local kubeconfig:

```sh
make run
```

### Build and Deploy

```sh
make docker-build IMG=<your-registry>/podgroup-controller:tag
make docker-push IMG=<your-registry>/podgroup-controller:tag
make deploy IMG=<your-registry>/podgroup-controller:tag
```

## Usage with KWOK

1. Deploy this controller to your cluster
2. Create a KWOK stage that waits for the ready label:

```yaml
apiVersion: kwok.x-k8s.io/v1alpha1
kind: Stage
metadata:
  name: pod-complete-on-ready
spec:
  resourceRef:
    apiGroup: v1
    kind: Pod
  selector:
    matchLabels:
      podgroup.jhwagner.github.io/ready: "true"
  next:
    statusTemplate: |
      {{ $now := Now }}
      conditions:
      - lastTransitionTime: {{ $now }}
        status: "True"
        type: Ready
      phase: Succeeded
```

3. Create pods with the podgroup label (e.g., via JobSet, Job, or StatefulSet)
4. The controller will mark them ready once all are Running, then KWOK will complete them

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
