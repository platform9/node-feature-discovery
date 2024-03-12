---
title: "NFD-Garbage-Collector"
layout: default
sort: 6
---

# NFD-GC
{: .no_toc}

---

NFD-GC (NFD Garbage-Collector) is preferably run as a Kubernetes deployment
with one replica. It makes sure that all
[NodeFeature](custom-resources.md#nodefeature) and
[NodeResourceTopology](custom-resources.md#noderesourcetopology) objects
have corresponding nodes and removes stale objects for non-existent nodes.

The daemon watches for Node deletion events and removes NodeFeature and
NodeResourceTopology objects upon them. It also runs periodically to make sure
no node delete event was missed and to remove any NodeFeature or
NodeResourceTopology objects that were created without corresponding node. The
default garbage collector interval is set to 1h which is the value when no
-gc-interval is specified.

## Configuration

In Helm deployments (see
[garbage collector parameters](../deployment/helm.md#garbage-collector-parameters))
NFD-GC will only be deployed when `enableNodeFeatureApi` or
`topologyUpdater.enable` is set to true.
