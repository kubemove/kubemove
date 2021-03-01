---
kep-number: 1
title: MoveEngine - High Level Design
authors:
  - "@vitta"
owners:
  - Uma Mukkara
  - Mayank
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
    * [Implementation Details/Notes/Constraints](#implementation-detailsnotesconstraints)
    * [Risks and Mitigations](#risks-and-mitigations)
* [Design acceptance criteria](#design-acceptance-criteria)
* [Implementation History](#implementation-history)

## Summary

From the KubeMove HLD KEP, KubeMoveEngine CRD defines the placeholders for the workflow for a given application. The CR is created by the user by invoking the spec that has the following details
   - Application to which this CR applies
   - Target location including the namespace. User can specify multiple target locations, in which case multiple KMovePair CRs get created.
   - The Dynamic Data Mobilizer or DDM details
   - schedule time to transfer data at regular interval
   - Telemetry management

`MoveEngine` controller that is watching this CR will deploy the KubeMove framework(`MoveEngine`, `DataSync` controller) at remote cluster to
setup the application and transfer the data to remote cluster. This controller creates the datasync CR periodically and invokes the `DataSync`
controller to initiate the data transfer through DDM. Once DDM completes the data transfer, `MoveEngine` controller updates the remote `DataSync` controller by creating dataSync CR at remote cluster to validate/verify the data transfer, and to perform the restore of data.

## Motivation

This KEP is to get into details of MoveEngine CR

### Goals

This KEP is to get into details of MoveEngine CR

### Non-Goals

## Proposal

`spec` of MoveEngine CR consists of following fields:
```
    # remoteCluster details
    movePair: cluster-2

    #namespaces which needs to be migrated, empty for default, all for all namespace
    namespace: ns1

    # targeted namespace, empty for source namespace
    remoteNamespace: ns2

    #label selectors
    selectors:
    - app.example.com=test

    # sync Period
    syncPeriod: */5 * * * *

    # moveEngine mode, value can be either Backup or Restore
    # Restore: if current engine is meant for restore purpose
    #         In case of restore, there should be only one moveEngine
    #         CR, with Restore mode, for given namespace and selectors.
    # Backup : if current engine is meant for backup purpose
    mode: Restore

    # name of the plugin to be invoked for data transfer of volume
    # this plugin will setup the necessary infrastructure ,
    # at source and/or remote, to transfer the app and/or data. 
    pluginProvider: PLUGIN_NAME

    # includeResources defines if volumes or cluster
    # resource needs to be synced or not
    includeResources: true
status:
    #syncing stage
    stage: Final

    #status of last sync
    lastStatus: Completed

    # last sync time
    lastSync: dd-mm-yyyy hh:mm:ss
    volumes:
    - namespace: ns1
      persistentVolumeClaim: claim1
      lastStatus: Completed
      lastSynced: dd-mm-yyyy hh:mm:ss
      reason: None
      volume: xyz
      remoteVolume: xyz1
      resources:
        kind: secret
        name: app-key
        phase: Created
        status: Completed
        reason: reason for failure
        lastSyncedTime: last synced time
```
Controller does the following:
1. Check the network connectivity to cluster in movePair CR

2. Discover the resources based on label selector
TODO: Fill the details regarding discovery of the resources

3. Deploy the resources on destination cluster
   - Create clientset of destination cluster
   - Transform the filtered resources. TODO: Fill more details
   - From source cluster, deploy transformed resources using above clientset in destination cluster

Another approach is to
 - Install object store in destination cluster
 - Send the filtered resources to object storage
 - Transform the resources in object store at destination
 - From destination cluster, deploy transformed resources

### Implementation Details/Notes/Constraints

TODO

### Risks and Mitigation


## Design acceptance criteria

This KEP is in provisional state. To move the design into `accepted` state, at least two sponsors are needed as design approvers from two different companies.

## Implementation History

- None


