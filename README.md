# CXI Kubernetes device plugin

## Introduction

This is a device plugin implementation that enables the registration of AMD GPU in a container cluster for compute workload.

## Deployment

Beta-version image available at `hub.docker.hpecorp.net/caio.davi/cxi-k8s-device-plugin:0.1`. 
The device plugin should run in every node with available Cassini NICs, in other to make them availabel to the K9s cluster. The easiest way of doing so is to create a Kubernetes DaemonSet:

```
kubectl create -f https://raw.github.hpe.com/caio-davi/cxi-k8s-device-plugin/main/hpecxi-device-plugin-ds.yaml?token=GHSAT0AAAAAAAADTAQ3T3L27CUNOOZPJWEGZ3AWJQQ
```

> Image only available at `hub.docker.hpecorp.net` for now. Make sure to login prior to run the command. 
