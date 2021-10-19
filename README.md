⛔️ DEPRECATED: This repository is no longer maintained.

# eth-header-sync

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/eth-header-sync)](https://goreportcard.com/report/github.com/vulcanize/eth-header-sync)

> Tool for syncing Ethereum headers into a Postgres database

## Table of Contents 
1. [Background](#background)
1. [Install](#install)
1. [Usage](#usage)
1. [Testing](#testing)
1. [Contributing](#contributing)
1. [License](#license)

## Background
Ethereum data is natively stored in key-value databases such as leveldb (geth) and rocksdb (openethereum).
Storage of Ethereum in these KV databases is optimized for scalability and ideal for the purposes of consensus,
it is not necessarily ideal for use by external applications. Moving Ethereum data into a relational database can provide
many advantages for downstream data consumers.

eth-header-sync validates and syncs Ethereum headers into Postgres. It syncs headers from both tail and head, at the head it maintains a validation window
(default size of 15) to handle chain reorgs.

This is useful when you want a minimal baseline from which to track and hash-link targeted data on the blockchain (e.g. individual smart contract storage values or event logs).
Examples of this usage are [eth-contract-watcher](https://github.com/vulcanize/eth-contract-watcher) and [eth-account-watcher](https://github.com/vulcanize/account_transformers).

Headers are fetched from the standard `eth_getBlockByNumber` JSON-RPC endpoint.


## Install

1. [Dependencies](#dependencies)
1. [Building the project](#building-the-project)
1. [Setting up the database](#setting-up-the-database)

### Dependencies
 - Go 1.12+
 - Postgres 11.2
 - Ethereum Node
   - [Go Ethereum](https://github.com/ethereum/go-ethereum/releases) (1.8.23+)
   - [Open Ethereum](https://github.com/openethereum/openethereum/releases) (1.8.11+)

### Building the project
Download the codebase to your local `GOPATH` via:

`go get github.com/vulcanize/eth-header-sync`

Move to the project directory:

`cd $GOPATH/src/github.com/vulcanize/eth-header-sync`

Be sure you have enabled Go Modules (`export GO111MODULE=on`), and build the executable with:

`make build`

If you need to use a different dependency than what is currently defined in `go.mod`, it may helpful to look into [the replace directive](https://github.com/golang/go/wiki/Modules#when-should-i-use-the-replace-directive).
This instruction enables you to point at a fork or the local filesystem for dependency resolution.

If you are running into issues at this stage, ensure that `GOPATH` is defined in your shell.
If necessary, `GOPATH` can be set in `~/.bashrc` or `~/.bash_profile`, depending upon your system.
It can be additionally helpful to add `$GOPATH/bin` to your shell's `$PATH`.

### Setting up the database
1. Install Postgres
1. Create a superuser for yourself and make sure `psql --list` works without prompting for a password.
1. `createdb vulcanize_public`
1. `cd $GOPATH/src/github.com/vulcanize/eth-header-sync`
1.  Run the migrations: `make migrate HOST_NAME=localhost NAME=vulcanize_public PORT=5432`
    - There are optional vars `USER=username` and `PASS=password` if the database user is not the default user `postgres` and/or a password is present
    - To rollback a single step: `make rollback NAME=vulcanize_public`
    - To rollback to a certain migration: `make rollback_to MIGRATION=n NAME=vulcanize_public`
    - To see status of migrations: `make migration_status NAME=vulcanize_public`

    * See below for configuring additional environments
    
In some cases (such as recent Ubuntu systems), it may be necessary to overcome failures of password authentication from
localhost. To allow access on Ubuntu, set localhost connections via hostname, ipv4, and ipv6 from peer/md5 to trust in: /etc/postgresql/<version>/pg_hba.conf

(It should be noted that trusted auth should only be enabled on systems without sensitive data in them: development and local test databases)

## Usage
`./eth-header-sync sync --config <config.toml> --starting-block-number <block-number>`

The config file must be formatted as follows, and should contain an RPC path to a running Ethereum full node:

```toml
[database]
    name     = "vulcanize_public"
    hostname = "localhost"
    user     = "postgres"
    password = ""
    port     = 5432

[client]
    rpcPath  = "http://127.0.0.1:8545"
```

### Testing
- Replace the empty `rpcPath` in the `environments/testing.toml` with a path to a full node's eth_jsonrpc endpoint (e.g. local geth node ipc path or infura url)
    - Note: must be mainnet
    - Note: integration tests require configuration with an archival node
- `make test` will run the unit tests and skip the integration tests
- `make integrationtest` will run just the integration tests
- `make test` and `make integrationtest` setup a clean `vulcanize_testing` db

## Maintainers
@vulcanize
@AFDudley
@i-norden


## Contributing
Contributions are welcome!

VulcanizeDB follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/1/4/code-of-conduct).


## License
[AGPL-3.0](LICENSE) © Vulcanize Inc
