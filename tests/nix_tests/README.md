# Nix Integration tests

This integration test suite uses nix for reproducible and configurable
builds allowing to run test calls using the Python `web3` library against
different evmOS and [Geth](https://github.com/ethereum/go-ethereum) clients
with multiple configurations.

## Installation

Refer to the corresponding [Nix installation guide](https://nix.dev/manual/nix/stable/installation/installing-binary.html)
for your respective platform to install Nix.

## Run Local

Run the following Makefile target,
which builds the corresponding Nix shell and then runs the tests.
When running for the first time, this will take a while to build the shell.

```
make run-nix-tests
```

Once you've run them once and, you can run:

```
nix-shell tests/nix_tests/shell.nix
cd tests/nix_tests
pytest -s -vv
```

If you're changing anything on the chain binary,
the first command will need to be run again.

## Caching

You can enable Binary Cache to speed up the tests:

```
nix-env -iA cachix -f https://cachix.org/api/v1/install
cachix use evmos
```
