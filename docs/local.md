# local test

### Requirements
- k3d
- skaffold
- ko
- kubectl

### How to use

Create a k3d registry + cluster
```
k3d cluster create --config k3d.yaml
skaffold dev
```