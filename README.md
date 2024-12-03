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
