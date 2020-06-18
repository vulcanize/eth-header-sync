# eth-header-sync

[![Go Report Card](https://goreportcard.com/badge/github.com/vulcanize/eth-header-sync)](https://goreportcard.com/report/github.com/vulcanize/eth-header-sync)

> Tool for syncing Ethereum headers into a Postgres database

## Table of Contents 
1. [Background](#background)
1. [Install](#install)
1. [Usage](#usage)
1. [Contributing](#contributing)
1. [License](#license)

## Background
Ethereum data is natively stored in key-value databases such as leveldb (geth) and rocksdb (openethereum).
Storage of Ethereum in these KV databases is optimized for scalability and ideal for the purposes of consensus,
it is not necessarily ideal for use by external applications. Moving Ethereum data into a relational database can provide
many advantages for downstream data consumers.

This tool syncs Ethereum headers in Postgres. Addtionally, it validates headers from the last 15 blocks to ensure the data is up to date and
handles chain reorgs by validating the most recent blocks' hashes and upserting invalid header records.

This is useful when you want a minimal baseline from which to track and hash-link targeted data on the blockchain (e.g. individual smart contract storage values or event logs).
Some examples of this are the [eth-contract-watcher]() and [eth-account-watcher]().

Headers are fetched by RPC queries to the standard `eth_getBlockByNumber` JSON-RPC endpoint, headers can be synced from anything
that supports this endpoint.


## Install

1. [Dependencies](#dependencies)
1. [Building the project](#building-the-project)
1. [Setting up the database](#setting-up-the-database)
1. [Configuring a synced Ethereum node](#configuring-a-synced-ethereum-node)

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
    - There is an optional var `USER=username` if the database user is not the default user `postgres`
    - To rollback a single step: `make rollback NAME=vulcanize_public`
    - To rollback to a certain migration: `make rollback_to MIGRATION=n NAME=vulcanize_public`
    - To see status of migrations: `make migration_status NAME=vulcanize_public`

    * See below for configuring additional environments
    
In some cases (such as recent Ubuntu systems), it may be necessary to overcome failures of password authentication from
localhost. To allow access on Ubuntu, set localhost connections via hostname, ipv4, and ipv6 from peer/md5 to trust in: /etc/postgresql/<version>/pg_hba.conf

(It should be noted that trusted auth should only be enabled on systems without sensitive data in them: development and local test databases)

### Configuring a synced Ethereum node
- To use a local Ethereum node, copy `environments/public.toml.example` to
  `environments/public.toml` and update the `rpcPath` in the config file.
  - `rpcPath` should match the local node's IPC filepath:
      - For Geth:
        - The IPC file is called `geth.ipc`.
        - The geth IPC file path is printed to the console when you start geth.
        - The default location is:
          - Mac: `<full home path>/Library/Ethereum/geth.ipc`
          - Linux: `<full home path>/ethereum/geth.ipc`
        - Note: the geth.ipc file may not exist until you've started the geth process
        - The default localhost HTTP URL is "http://127.0.0.1:8545"

      - For OpenEthereum:
        - The IPC file is called `jsonrpc.ipc`.
        - The default location is:
          - Mac: `<full home path>/Library/Application\ Support/io.parity.ethereum/`
          - Linux: `<full home path>/local/share/io.parity.ethereum/`

- To use a remote Ethereum node, simply set the `rpcPath` in the config file to the HTTP RPC endpoint url for the remote node
    - The default HTTP port # for Geth and OpenEthereum is 8545

## Usage
`./eth-header-sync sync --config <config.toml> --starting-block-number <block-number>`

The config file must be formatted as follows, and should contain an RPC path to a running Ethereum node:

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


## Maintainers
@vulcanize
@AFDudley
@i-norden



## Contributing
Contributions are welcome!

VulcanizeDB follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/1/4/code-of-conduct).


## License
[AGPL-3.0](LICENSE) © Vulcanize Inc