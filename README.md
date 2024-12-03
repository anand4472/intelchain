# Intelchain

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
### **Docker** (for testing)

On macOS:
```bash
brew install --cask docker
open /Applications/Docker.app
```
On Linux, reference official documentation [here](https://docs.docker.com/engine/install/).
### **Bash 4+**

For macOS, you can reference this [guide](http://tldrdevnotes.com/bash-upgrade-3-4-macos). For Linux, you can reference this [guide](https://fossbytes.com/installing-gnu-bash-4-4-linux-distros/).

## Dev Environment


### First Install
Clone and set up all of the repos with the following set of commands:

1. Create the appropriate directories:
```bash
mkdir -p $(go env GOPATH)/src/github.com/intelchain-itc
cd $(go env GOPATH)/src/github.com/intelchain-itc
```
> If you get 'unknown command' or something along those lines, make sure to install [golang](https://golang.org/doc/install) first.

2. Clone this repo & dependent repos.
```bash
git clone https://github.com/intelchain-itc/mcl.git
git clone https://github.com/intelchain-itc/bls.git
git clone https://github.com/intelchain-itc/intelchain.git
cd intelchain
```
