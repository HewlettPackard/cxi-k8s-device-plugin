# CXI Kubernetes device plugin

## Introduction

This is a device plugin implementation that enables the registration of HPE Slingshot Cassini NICs in a Kubernetes cluster for compute workload.

## Requirements

##### - Kubernetes v1.30 or higher

##### - CXI Services Enabled

This step must be executed for every node and every Slingshot NIC that servers the Kubernetes cluster:
```
cxi_service enable -d hsnX -s 1
```

##### - Enable Mutating Admission Policy

Mutating Admission Policy is an alpha feature released in Kubernetes v1.30. As with any alpha feature, it must be explicitly enabled before usage. Add the following lines to /etc/kubernetes/manifests/kube-apiserver.yaml:
```
- --feature-gates=MutatingAdmissionPolicy=true
- --runtime-config=admissionregistration.k8s.io/v1alpha1=true
```

##### - Multus CNI 

Since we need multiple NICs, we are going to use Multus CNI (a CNI meta-plugin) that enables attaching multiple network interfaces to pods.
```
kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset-thick.yml
```

##### - Whereabouts CNI

Whereabouts is an IP Address Management (IPAM) CNI plugin that assigns IP addresses cluster wide.  
```
git clone https://github.com/k8snetworkplumbingwg/whereabouts
cd whereabouts
kubectl apply \
    -f doc/crds/daemonset-install.yaml \
    -f doc/crds/whereabouts.cni.cncf.io_ippools.yaml \
    -f doc/crds/whereabouts.cni.cncf.io_overlappingrangeipreservations.yaml
```
<!-- ##### Network Attachment Definitions

We need to create Custom Resource Definitions (CRDs) describing the configurations for each of the additional interfaces that we want to attach to the pods.
```
kubectl apply -f ./deploy/NAD
``` -->
## Deployment

Alpha-version image available at `hub.docker.hpecorp.net/caio.davi/cxi-k8s-device-plugin:0.1`. 
The device plugin should run in every node with available Cassini NICs, in other to make them availabel to the K8s cluster. The easiest way of doing so is to create a Kubernetes DaemonSet:

```
kubectl apply \
    -f ./deploy/NetworkAttachmentDefinition \
    -f ./deploy/MutatingAdmissionPolicy \ 
    -f ./deploy/hpecxi-device-plugin-ds.yaml
```

> Image only available at `hub.docker.hpecorp.net` for now. Make sure to login prior to run the command. 
