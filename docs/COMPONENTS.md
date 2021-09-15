# Resources

## Quarantine

The quarantine resource represents a lifecycle management for isolating nodes and pods as well. Creating a resource is similar to start a quarantine and deleting it is similar to stopping it. 

Here is an overview of the reconciling process:

![Alt text](img/workflow.png?raw=true "Overview")
### debug

You can configure if da debug pod is deployed on affected nodes. It's also possible to an other image than the default.
### nodes

There are configuration options per node. This contains workload which pods should be isolated or not rescheduled, using a specific debug pod for a node and adding taint to a node. Workloads which are configured to be isolated are merged with configured resources under .spec.resources.
### resources

This is a list of workloads whose pods should be isolated on each affected node configured under .spec.nodes and is merged with node specific configurations.
