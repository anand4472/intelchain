# Intelchain


## General Documentation

https://docs.intelchain.org

## API Guide

http://api.intelchain.org/

## Requirements

### **Go 1.22.5**
### **GMP and OpenSSL**

On macOS:
```bash
brew install gmp
brew install openssl
sudo ln -sf /usr/local/opt/openssl@1.1 /usr/local/opt/openssl
```
On Linux (Ubuntu)
```bash
sudo apt install libgmp-dev  libssl-dev  make gcc g++
```
On Linux (Cent OS / Amazon Linux 2)
```bash
sudo yum install glibc-static gmp-devel gmp-static openssl-libs openssl-static gcc-c++
```

### First Install
Clone and set up all of the repos with the following set of commands:

1. Create the appropriate directories:
```bash
mkdir -p $(go env GOPATH)/src/github.com/zennittians
cd $(go env GOPATH)/src/github.com/zennittians
```
> If you get 'unknown command' or something along those lines, make sure to install [golang](https://golang.org/doc/install) first.

2. Clone this repo & dependent repos.
```bash
git clone https://github.com/zennittians/mcl.git
git clone https://github.com/zennittians/bls.git
git clone https://github.com/zennittians/intelchain.git
cd intelchain
```

3. Build the intelchain binary & dependent libs
```
go mod tidy
make
```
> Run `bash scripts/install_build_tools.sh` to ensure build tools are of correct versions.
> If you get 'missing go.sum entry for module providing package <package_name>', run `go mod tidy`.


## Build

The `make` command should automatically build the intelchain binary & all dependent libs.

However, if you wish to bypass the Makefile, first export the build flags:
```bash
export CGO_CFLAGS="-I$GOPATH/src/github.com/zennittians/bls/include -I$GOPATH/src/github.com/zennittians/mcl/include -I/opt/homebrew/opt/openssl@1.1/include"
export CGO_LDFLAGS="-L$GOPATH/src/github.com/zennittians/bls/lib -L/opt/homebrew/opt/openssl@1.1/lib"
export LD_LIBRARY_PATH=$GOPATH/src/github.com/zennittians/bls/lib:$GOPATH/src/github.com/zennittians/mcl/lib:/opt/homebrew/opt/openssl@1.1/lib
export LIBRARY_PATH=$LD_LIBRARY_PATH
export DYLD_FALLBACK_LIBRARY_PATH=$LD_LIBRARY_PATH
export GO111MODULE=on
```

Then you can build all executables with the following command:
```bash
bash ./scripts/go_executable_build.sh -S
```
> Reference `bash ./scripts/go_executable_build.sh -h` for more build options

## Debugging

One can start a local network (a.k.a localnet) with your current code using the following command:
```bash
make debug
```
> This localnet has 2 shards, with 11 nodes on shard 0 (+1 explorer node) and 10 nodes on shard 0 (+1 explorer node).
>
> The shard 0 endpoint will be on the explorer at `http://localhost:9599`. The shard 1 endpoint will be on the explorer at `http://localhost:9598`.
>
> You can view the localnet configuration at `/test/configs/local-resharding.txt`. The fields for the config are (space-delimited & in order) `ip`, `port`, `mode`, `bls_pub_key`, and `shard` (optional).

One can force kill the local network with the following command:
```bash
make debug-kill
```
> You can view all make commands with `make help`

## Testing

To keep things consistent, we have a docker image to run all tests. **These are the same tests ran on the pull request checks**.

Note that all test Docker containers bind several ports to the host machine for your convenience. The ports are:
* `9500` - Shard 0 RPC for a validator
* `9501` - Shard 1 RPC for a validator
* `9599` - Shard 0 RPC for an explorer
* `9598` - Shard 1 RPC for an explorer
* `9799` - Shard 0 Rosetta (for an explorer)
* `9798` - Shard 1 Rosetta (for an explorer)
* `9899` - Shard 0 WS for an explorer
* `9898` - Shard 1 WS for an explorer
> This allows you to use curl, itc CLI, postman, rosetta-cli, etc... on your host machine to play with or probe the localnet that was used for the test.

### Go tests
To run this test, do:
```bash
make test-go
```
This test runs the go tests along with go lint, go fmt, go imports, go mod, and go generate checks.

### RPC tests
To run this test, do:
```bash
make test-rpc
```

If you wish to debug further with the localnet after the tests are done, open a new shell and run:
```bash
make test-rpc-attach
```

### Rosetta tests
To run this test, do:
```bash
make test-rosetta
```

Similar to the RPC tests, if you wish to debug further with the localnet after the tests are done, open a new shell and run:
```bash
make test-rosetta-attach
```
