---
kep-number: 1
title: KubeMove - High Level Design
authors:
  - "@umamukkara"
owners:
  - Uma Mukkara
editor: 
creation-date: 2019-05-16
last-updated: 2019-05-19
status: provisional
see-also:
  - KEP-None
replaces:
  - KEP-None
superseded-by:
  - KEP-None
---

# KubeMove - High Level Design



## Table of Contents

* [Summary](#summary)
* [Motivation](#motivation)
    * [Goals](#goals)
    * [Non-Goals](#non-goals)
* [Proposal](#proposal)
    * [User Stories](#user-stories-optional)
    * [Implementation Details/Notes/Constraints [optional]](#implementation-detailsnotesconstraints-optional)
    * [Risks and Mitigations](#risks-and-mitigations)
* [Graduation Criteria](#graduation-criteria)
* [Implementation History](#implementation-history)

## Summary

Mobility of applications is a need that often arises because of many reasons. Developers need an easy way to initiate the mobility, monitor the progress and manage the entire process. This KEP is a proposal of the architecture and design of the KubeMove project which provides an operator and a set of APIs to achive the required mobility of the applications. KubeMove also proposes a CLI using which developers can invoke the APIs easily and control the mobility workflow. The API set itself is an initial proposal which we expect to evolve during the community collaboaration process.

KubeMove proposes an initial set of FIVE APIs and an interface called DDM (Dynamic Data Mobilizer) to implement either application specific or storage specific data mobilizer using which the actual data mobilty happens. In this KEP, we cover the potential list of use cases for KubeMove.

Proposed kubemove apis/CRDs are :

- moveengine.kubemove.io

- movepair.kubemove.io

- datasync.kubemove.io

- moveswitch.kubemove.io

- movereverse.kubemove.io

  

movepair invokes the Dynamic Data Mobilizer (DDM), which implements the API datasync.kubemove.io



## Motivation

There are tools available for backup and restore. Kubernetes is also working on the APIs for backup and restore. However, there is a need to orchestrate the entire moblity	flow across provider boundaries in order to get the applications moved between the user's cluster which may be across the clouds or hybrid cloud environments. Cloud provider or a middleware has to coordinate the data mobility between the end points. The motivation for this project is to identify and collaborate such API with the interested providers and middleware providers.

### Goals

- Identify the mobility workflow with provider or middleware software in the middle
- Identify the interfaces to invoke and monitor the mobility process by the end users such as application developers
- Identify the list of CLIs the users can use to move and monitor the applications

### Non-Goals

- Networking connectivity between the two end points. This is done external to KubeMove. There are new interesting projects like [Rancher Submariner](https://github.com/rancher/submariner) which may be adopted to handle the connections between the source and destination
- Authentication and authorisation. This is also done external to KubeMove. Technologies like Istio, Envoy, Federation will be part of the authentication and authorisation. 



## Proposal

User treats application mobility as a need to move the application from one place to another rather than treating the process as a backup and restore. For this reason, there are two types of interfaces that are being proposed.

1. APIs for the end user to interact with the application and the data mobilizer
2. APIs for the cloud provider or the middleware to control the data mobility flow where they can also control authentication and authorization when and if needed

There are five inital APIs that are being proposed:

1. **KubeMoveEngine:** This CRD defines the placeholders for the workflow for a given application. The CR is created either by the user by invoking the spec that has the following details
   - Application to which this CR applies
   - Target location including the namespace. User can specify multiple target locations, in which case multiple KMovePair CRs get created.
   - The Dynamic Data Mobilizer or DDM details
   - One-time offline move or live movement at a future date with continuous data transfer starting immediately
   - Telemetry management
2. **KubeMovePair:** This CRD defines the data channel between the source and destination endpoints. KubeMovePair CR is created automatically by KubeMoveEngine controller. As indicated in the non-goals section of this document, setting up of the actual communication channel is not in the scope of KubeMove, it is done external to KubeMove. KubeMovePair contains the status of the data channel (init, in-progress, active, complete, unknown). KubeMovePair controller that is watching this CR will manage the data channel in accordance with the Kubernetes controller pattern of "desired state vs current state". This controller invokes DDM and waits for the DataSync CR to be created. Application developers or application vendors or cloud providers or storage providers may write their own DDM implementations to handle the data sync between source and target end points.
3. **DataSync:** This CRD defines the interface that DDM implementors have to expose. At the end of a succesful data channel creation between the two end points, DataSync CR is created by the DDM. This CR provides granular status of the data channel. DDM is either application specific or storage vendor specific. For example, a PostgreSQL DDM may establish the data channel on top of a KubeMovePair and manage the async snapshots to be transferred at regular interval from source to destination.
4. **KubeMoveSwitch:** This CRD is defined for completing the application mobility part, from source to target. The controller watching the CR handles the flow of bringing down the source application and enabling the target application with the required spec. If any change to the application that needs to be applied on the target side, that change is handled through this CRD. The target side switch hooks provide post switch execution scripts like changing the ingress or istio configuration etc.
5. **KubeMoveReverse**: This is an optional step. Many times, the application may need to be moved back and forth between two destinations, for example in scenarios where blue-green deployment strategy is followed. If an application has to be moved back to the original location, the data sync direction needs to be changed, which means that an entire data channel has to be setup in the reverse direction. Using this CR, the controller manages the establishment of the channel till the DataSync CR is created for the reverse data transfer.



**KubeMove interaction with the end user:**

End user invokes the data mobility flow by inserting the annotation `kubemove.io/kubemoveapply="true"`. In addition, the user also identifies the name of the KubeMoveEngine that handles the mobility by using the annotation `kubemove.io/engine="myBusyBoxKMEngine"` . These two annotations together provide an entry point of KubeMove to the application moblity. User then invoke the KubeMoveEngine to specify the details of required application mobility.

**KubeMove interaction with cloud provider:**

Cloud provider may interact with KubeMovePair for monitoring the status and invoke KubeMoveSwitch either automatically or by interacting with the user.  Cloud provider can also participate in the manipulation of KubeMoveSwitch and KubeMoveReverse CRs.



### User Stories

- Hybrid and multi-cloud deployments - Move applications on-prem to cloud or vice-versa or from cloud-cloud
- Onramp to Kubernetes - KubeMove can help migrate the data from legacy volumes onto Kubernetes
- Kubernetes and/or application upgrades - Applications may need to be moved back and forth while following blue green strategy

### Implementation Details/Notes/Constraints

KubeMove project defines the spec for the CRDs and implements the controllers to handle the end-to-end application mobility. We also plan to implement a reference DDM that handles the data sync between source and target using Restik. 

### Risks and Mitigations

Establishing the secure communication between the end points across boundaries can pose as a challenge. The actual secure channel establishment is not in the scope of KubeMove but needs to be given a complete thought for data security on the wire. 

## Graduation Criteria

This KEP is in proposal state. To graduate, at least two sponsors are needed as design approvers from two different companies. 



## Implementation History

- None





