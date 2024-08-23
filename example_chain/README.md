# Example evmOS Chain

This directory contains an example chain that uses the evmOS modules.
It is based on the simapp implementation on the Cosmos SDK repository,
which itself is a simplified version of a Cosmos SDK-based blockchain.

This chain implementation is used to demonstrate the integration of evmOS
as well as to provide a chain object for testing purposes within the repository.

## Config

By default, this chain has the following configuration:

| Option              | Value                  |
|---------------------|------------------------|
| Binary              | `osd`                  |
| Chain ID            | `os_9005-1`            |
| Custom Opcodes      | -                      |
| Default Token Pairs | 1 for the native token |
| Denomination        | `aevmos`               |
| EVM flavor          | permissionless         |
| Enabled Precompiles | all                    |

## Running The Chain

To run the example, execute the local node script found within this repository:

```bash
./local_node.sh [FLAGS]
```

Available flags are:

- `-y`: Overwrite previous database
- `-n`: Do **not** overwrite previous database
- `--no-install`: Skip installation of the binary

## Available Cosmos SDK Modules

As mentioned above, this exemplary chain implementation is a reduced version of `simapp`.
Specifically, instead of offering access to all Cosmos SDK modules, it just includes the following:

- `auth`
- `authz`
- `bank`
- `capability`
- `consensus`
- `distribution`
- `evidence`
- `feegrant`
- `genutil`
- `gov`
- `mint`
- `params`
- `slashing`
- `staking`
- `upgrade`
