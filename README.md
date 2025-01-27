# CXI Kubernetes device plugin

## Introduction

This is a device plugin implementation that enables the registration of HPE Slingshot Cassini NICs in a Kubernetes cluster for compute workload.

## Deployment

Alpha-version image available at `hub.docker.hpecorp.net/caio.davi/cxi-k8s-device-plugin:0.1`. 
The device plugin should run in every node with available Cassini NICs, in other to make them availabel to the K8s cluster. The easiest way of doing so is to create a Kubernetes DaemonSet:

```
kubectl create -f https://raw.github.hpe.com/caio-davi/cxi-k8s-device-plugin/main/deploy/hpecxi-device-plugin-ds.yaml
```

> Image only available at `hub.docker.hpecorp.net` for now. Make sure to login prior to run the command. 
