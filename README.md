# CXI Kubernetes device plugin

> [!CAUTION]
This is a beta software, not recommended for production systems.

## Table of Contents

- [CXI Kubernetes device plugin](#cxi-kubernetes-device-plugin)
  - [Table of Contents](#table-of-contents)
  - [Introduction](#introduction)
    - [Requirements](#requirements)
      - [Kubernetes v1.30 or higher](#kubernetes-v130-or-higher)
      - [CXI Services Enabled](#cxi-services-enabled)
      - [Enable Mutating Admission Policy](#enable-mutating-admission-policy)
      - [Multus CNI](#multus-cni)
      - [Whereabouts CNI](#whereabouts-cni)
    - [Build](#build)
    - [Deployment](#deployment)
    - [Running HPE Slingshot Jobs](#running-hpe-slingshot-jobs)
  - [CXI CDI Generator](#cxi-cdi-generator)
    - [Usage](#usage)
    - [Environment Variables](#environment-variables)

## Introduction

This device plugin implementation enables the registration of HPE Slingshot Cassini NICs in a Kubernetes cluster for compute workloads.

### Requirements

#### Kubernetes v1.30 or higher

#### CXI Services Enabled

This step must be executed for every node and every Slingshot NIC that serves the Kubernetes cluster:
```
cxi_service enable -d cxiX -s 1
```

#### Enable Mutating Admission Policy

Mutating Admission Policy is an alpha feature released in Kubernetes v1.30. As with any alpha feature, it must be explicitly enabled before usage. Add the following lines to /etc/kubernetes/manifests/kube-apiserver.yaml:
```
- --feature-gates=MutatingAdmissionPolicy=true
- --runtime-config=admissionregistration.k8s.io/v1alpha1=true
```

#### Multus CNI 

Since it requires multiple NICs, we use Multus CNI (a CNI meta-plugin) to enable attaching multiple network interfaces to pods.
```
kubectl apply -f https://raw.githubusercontent.com/k8snetworkplumbingwg/multus-cni/master/deployments/multus-daemonset-thick.yml
```

#### Whereabouts CNI

Whereabouts is an IP Address Management (IPAM) CNI plugin that assigns IP addresses cluster-wide.  
```
git clone https://github.com/k8snetworkplumbingwg/whereabouts
cd whereabouts
kubectl apply \
    -f doc/crds/daemonset-install.yaml \
    -f doc/crds/whereabouts.cni.cncf.io_ippools.yaml \
    -f doc/crds/whereabouts.cni.cncf.io_overlappingrangeipreservations.yaml
```

### Build

Just build it from the Makefile:
```
make build
```

It will create both the `cxi-k8s-device-plugin` and `cxi-cdi-generator` in the `./bin` directory.

### Deployment

Alpha-version image available at `ghcr.io/hewlettpackard/cxi-k8s-device-plugin:0.0.1-beta`. 
The device plugin should run in every node with available Cassini NICs, in order to make them available to the K8s cluster. The easiest way of doing so is to create a Kubernetes DaemonSet:

```
kubectl apply \
    -f ./deploy/NetworkAttachmentDefinition \
    -f ./deploy/MutatingAdmissionPolicy \ 
    -f ./deploy/hpecxi-device-plugin-ds.yaml
```

> [!IMPORTANT]
> Make sure the IPAM definitions in the `./deploy/NetworkAttachmentDefinition` are following your cluster network requirements.

### Running HPE Slingshot Jobs

With the daemonset deployed, HPE Slingshot NICs can now be requested by a container using the `beta.hpe.com/cxi` resource type:

```shell
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Pod
metadata:
  name: cxi-test
  labels:
    app: test
spec:
  containers:
    - name: cxi-test-container
      image: nicolaka/netshoot
      command:
        - sleep
        - "3600"
      resources:
        requests:
          beta.hpe.com/cxi: 4 # requesting four NICs
        limits:
          beta.hpe.com/cxi: 4 # requesting four NICs
EOF
```

## CXI CDI Generator

The cxi-cdi-generator is a command-line tool to generate [Container Device Interface (CDI)](https://github.com/cncf-tags/container-device-interface/) specifications for HPE Slingshot NIC devices. 

### Usage

The built executable `cxi-cdi-generator` will discover the `cxi` devices in your system and generate the CDI profile in `/etc/cdi`.
```
cxi-cdi-generator
```

This usually requires admin privileges, since `/etc` is a restricted directory. Alternatively, you can also save in a directory of your choice:
```
cxi-cdi-generator --cdi-dir {dir_path}
```

You may also use `cxi-cdi-generator --help` for more options. 

### Environment Variables

During execution, `cxi-cdi-generator` will look into the following default paths for the discovery process. If any of these components are placed in a different path in your system, you can set it up using the corresponding environment variable.  

| EnvVar     | Default Value     | Details               |
| ---------- | ----------------- | --------------------- |
| SYSFS_ROOT | `/sys`            | SystemFs path on host |
| DEVFS_ROOT | `/dev`            | DevFs path on host    |
| OFI_ROOT   | `/opt/cray/lib64` | Libfabric pathon host |
| CXI_ROOT   | `/usr/lib64`      | Libcxi path on host  |
