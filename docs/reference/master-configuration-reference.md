---
title: "Master config reference"
layout: default
sort: 3
---

# Configuration file reference of nfd-master
{: .no_toc}

## Table of contents
{: .no_toc .text-delta}

1. TOC
{:toc}

---

See the
[sample configuration file](https://github.com/kubernetes-sigs/node-feature-discovery/blob/{{site.release}}/deployment/components/master-config/nfd-master.conf.example)
for a full example configuration.

## noPublish

`noPublish` option disables updates to the Node objects in the Kubernetes
API server, making a "dry-run" flag for nfd-master. No Labels, Annotations, Taints
or ExtendedResources of nodes are updated.

Default: `false`

Example:

```yaml
noPublish: true
```

## extraLabelNs
`extraLabelNs` specifies a list of allowed feature
label namespaces. This option can be used to allow
other vendor or application specific namespaces for custom labels from the
local and custom feature sources, even though these labels were denied using
the `denyLabelNs` parameter.

The same namespace control and this option applies to Extended Resources (created
with `resourceLabels`), too.

Default: *empty*

Example:

```yaml
extraLabelNs: ["added.ns.io","added.kubernets.io"]
```

## denyLabelNs
`denyLabelNs` specifies a list of excluded
label namespaces. By default, nfd-master allows creating labels in all
namespaces, excluding `kubernetes.io` namespace and its sub-namespaces
(i.e. `*.kubernetes.io`). However, you should note that
`kubernetes.io` and its sub-namespaces are always denied.
This option can be used to exclude some vendors or application specific
namespaces.
Note that the namespaces `feature.node.kubernetes.io` and `profile.node.kubernetes.io`
and their sub-namespaces are always allowed and cannot be denied.

Default: *empty*

Example:

```yaml
denyLabelNs: ["denied.ns.io","denied.kubernetes.io"]
```

## resourceLabels

**DEPRECATED**: [NodeFeatureRule](../usage/custom-resources.md#nodefeaturerule)
should be used for managing extended resources in NFD.

The `resourceLabels` option specifies a list of features to be
advertised as extended resources instead of labels. Features that have integer
values can be published as Extended Resources by listing them in this option.

Default: *empty*

Example:

```yaml
resourceLabels: ["vendor-1.com/feature-1","vendor-2.io/feature-2"]
```

## enableTaints
`enableTaints` enables/disables node tainting feature of NFD.

Default: *false*

Example:

```yaml
enableTaints: true
```

## labelWhiteList
`labelWhiteList` specifies a regular expression for filtering feature
labels based on their name. Each label must match against the given reqular
expression in order to be published.

Note: The regular expression is only matches against the "basename" part of the
label, i.e. to the part of the name after '/'. The label namespace is omitted.

Default: *empty*

Example:

```yaml
labelWhiteList: "foo"
```

## resyncPeriod

The `resyncPeriod` option specifies the NFD API controller resync period.
The resync means nfd-master replaying all NodeFeature and NodeFeatureRule objects,
thus effectively re-syncing all nodes in the cluster (i.e. ensuring labels, annotations,
extended resources and taints are in place).
Only has effect when the [NodeFeature](../usage/custom-resources.md#nodefeature)
CRD API has been enabled with [`-enable-nodefeature-api`](master-commandline-reference.md#-enable-nodefeature-api).

Default: 1 hour.

Example:

```yaml
resyncPeriod: 2h
```

## leaderElection

The `leaderElection` section exposes configuration to tweak leader election.

### leaderElection.leaseDuration

`leaderElection.leaseDuration` is the duration that non-leader candidates will
wait to force acquire leadership. This is measured against time of
last observed ack.

A client needs to wait a full LeaseDuration without observing a change to
the record before it can attempt to take over. When all clients are
shutdown and a new set of clients are started with different names against
the same leader record, they must wait the full LeaseDuration before
attempting to acquire the lease. Thus LeaseDuration should be as short as
possible (within your tolerance for clock skew rate) to avoid a possible
long waits in the scenario.

Default: 15 seconds.

Example:

```yaml
leaderElection:
  leaseDurtation: 15s
```

### leaderElection.renewDeadline

`leaderElection.renewDeadline` is the duration that the acting master will retry
refreshing leadership before giving up.

This value has to be lower than leaseDuration and greater than retryPeriod*1.2.

Default: 10 seconds.

Example:

```yaml
leaderElection:
  renewDeadline: 10s
```

### leaderElection.retryPeriod

`leaderElection.retryPeriod` is the duration the LeaderElector clients should wait
between tries of actions.

It has to be greater than 0.

Default: 2 seconds.

Example:

```yaml
leaderElection:
  retryPeriod: 2s
```

### nfdApiParallelism

The `nfdApiParallelism` option can be used to specify the maximum
number of concurrent node updates.

It takes effect only when `-enable-nodefeature-api` has been set.

Default: 10

Example:

```yaml
nfdApiParallelism: 1
```
