<img src="logo/logo.png" alt="kube move logo" width="60%" />

[![Go Report Card](https://goreportcard.com/badge/github.com/kubemove/kubemove)](https://goreportcard.com/report/github.com/kubemove/kubemove)
[![Build Status](https://travis-ci.org/kubemove/kubemove.svg?branch=master)](https://travis-ci.org/kubemove/kubemove)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/2830/badge)](https://bestpractices.coreinfrastructure.org/projects/2830)

# Overview

Workloads running on Kubernetes do not make assumptions about the underlying infrastructure. However,  it is still non-trivial to move them across different clusters. It is also non trivial to move them to another namespace within the same cluster scope such as to a different zone. KubeMove provides a set of tools and API to coordinate the orchestration and workflow of moving an application in production across cluster or namespace boundaries. 



# Architecture/Design proposal

The initial KEP (KubeMove Enhancement Proposal) is at [KEP-1](<https://github.com/kubemove/kubemove/blob/master/keps/0001-kep-kubemove-hld.md>)

Join the discussion through the issue [here](<https://github.com/kubemove/kubemove/issues/14>)



# Using Kubemove

- Install kubemove using helm - `helm install kubemove.io/kubemove`
- Annotate the application to be moved with `kubemove.io/kubemove.enable="true"`
- Use `MoveEngine` API with the required configuration to start the data movement or sync
- Wait for the movement or datasync to complete  and switch the application on the target cluster using `MoveSwitch` API

# CLI

**Example move commands:**

`kubectl km move init <app> <kubemove-template.yaml>`

`kubectl km status <movepair-cr>`

`kubectl km move switch <movepair>`



# Use cases

- Hybrid and multi-cloud deployments - Move applications on-prem to cloud or vice-versa or from cloud-cloud
- Onramp to Kubernetes - KubeMove can help migrate the data from legacy volumes onto Kubernetes
- Kubernetes and/or application upgrades - Applications may need to be moved back and forth while following blue green strategy

# Developer Guide

Kubemove uses `Makefile` based build process. This section will give you an overview on how to build and deploy Kubemove resources.

## Setup Environment

Setup your development environment by the following steps.

- Use your own docker account for the docker images:
```bash
export REGISTRY=<your docker username>
```
- For AWS ECR

```bash
export REGISTRY=<accountid>.dkr.ecr.<Region>.amazonaws.com
```
- Changes in build/push

  Replace this `sudo docker login "${REGISTRY}" -u "${DNAME}" -p "${DPASS}";` with
  `aws ecr get-login-password --region <Region> | docker login --username AWS --password-stdin "${REGISTRY}"`

- Build a developer image with all dependencies:
```bash
make dev-image
```

## Code Generation

If you update any API or any gRPC protos, then generate respective codes by the following steps.

 - Generate gRPC codes:
 ```bash
make gen-grpc
 ```

- Generate CRDs and respective codes.
```bash
make gen-crds
```

- Update respective controllers
```bash
make gen-k8s
```

Alternatively, you can run the following command that will run all the previous code generation commands.

```bash
make gen
```

## Build the Binaries and Docker Images

If you update any codes, then re-build the project and the respective docker images.

- Run `gofmt`
```bash
make format
```

- Run linter
```bash
make lint
```

- Build binaries
```bash
make build
```

- Build docker images
```bash
make images
```

- Push docker images
```bash
make deploy-images
```

## Deploy Kubemove

At first, create two different clusters. Make sure `KUBECONFIG` environment variable is pointing to the right cluster config file.

- Specify the source cluster, and the destination cluster
```bash
export SRC_CONTEXT=<source cluster context>
export DST_CONTEXT=<destination cluster context>
```

- Register the Kubemove CRDs
```bash
make register_crds
```
- Create the RBAC resources
```bash
make create_rbac_resources
```

- Deploy MovePair controller
```bash
make deploy_mp_ctrl
```

- Deploy MoveEngine controller
```bash
make deploy_me_ctrl
```

- Deploy DataSync controller
```bash
make deploy_ds_ctrl
```

You can also install all the Kubemove resources by using the following command:
```bash
make deploy_kubemove
```

If you are using two local kind cluster, you can create a MovePair using the following command:
```bash
make create_local_mp
```

## Removing Kubemove

You can uninstall/remove all Kubemove resources created in the deploy section just by replacing
`deploy`,`create`, or `register` word of the respective command with `remove`. For example:
```bash
make remove_mp_ctrl
```

Alternatively, you can use the following command to remove all the Kubemove resources created in the clusters using
the following command,
```bash
make purge_kubemove
```

# License

KubeMove is developed under Apache 2.0 license.



# Contributing

We welcome participation from the community in defining more use cases, developing API spec and implementation. Please write new issues as you like.

If you would like to know the ongoing work and its status, you can join us at status [meeting](<https://meet.google.com/ueh-fycm-aex>) that happens every Monday / Wednesday / Friday at 2.30 PM IST.

Minutes of meeting are captured in this [doc](<https://docs.google.com/document/d/1B7y28-WUiOy_RnFVbGF59BDYOMYXV5cPTnfxBbcFXp4/edit?usp=sharing>)
