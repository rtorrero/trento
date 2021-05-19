# Tagging the systems

In order to group and filter the systems a tagging mechanism can be used. This tags are placed as
meta-data in the agent nodes. Find information about how to set meta-data in the agents at: https://www.consul.io/docs/agent/options#node_meta

As an example, check the [meta-data file](./examples/trento-config.json) file. This file must be
located in the folder set as `-config-dir` during the agent execution.

The next items are reserved:
- `trento-ha-cluster`: Cluster which the system belongs to
- `trento-sap-environment`: Environment in which the system is running
- `trento-sap-landscape`: Landscape in which the system is running
- `trento-sap-environment`: SAP system (composed by database and application) in which the system is running

## Setting the tags from the KV storage

These reserved tags can be automatically set and updated using the [consul-template](https://github.com/hashicorp/consul-template).
To achieve this, the tags information will come from the KV storage.

Set the metadata in the next paths:
- `trento/nodename/metadata/ha-cluster`
- `trento/nodename/metadata/sap-environment`
- `trento/nodename/metadata/sap-landscape`
- `trento/nodename/metadata/sap-system`

Notice that a new entry must exists for every node.

`consul-template` starts directly with the `trento` agent. It provides some configuration options to synchronize the utility with the consul agent.

- `config-dir`: Consul agent configuration files directory. It must be the same used by the consul agent. The `trento` agent creates a new folder with the node name where the trento meta-data configuration file is stored (e.g. `consul.d/node1/trento-config.json`).
- `consul-template`: Template used to populate the trento meta-data configuration file (by default [meta-data file][./examples/trento-config.json] is used).
