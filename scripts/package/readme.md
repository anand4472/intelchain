# Introduction
This document introduces the Intelchain's package release using standard packaging system, RPM and Deb packages.

Standard packaging system has many benefits, like extensive tooling, documentation, portability, and complete design to handle different situation.

# Package Content
The RPM/Deb packages will install the following files/binary in your system.
* /usr/sbin/intelchain
* /usr/sbin/intelchain-setup.sh
* /usr/sbin/intelchain-rclone.sh
* /etc/intelchain/intelchain.conf
* /etc/intelchain/rclone.conf
* /etc/systemd/system/intelchain.service
* /etc/sysctl.d/99-intelchain.conf

The package will create `intelchain` group and `intelchain` user on your system.
The intelchain process will be run as `intelchain` user.
The default blockchain DBs are stored in `/home/intelchain/intelchain_db_?` directory.
The configuration of intelchain process is in `/etc/intelchain/intelchain.conf`.

# Package Manager
Please take some time to learn about the package managers used on Fedora/Debian based distributions.
There are many other package managers can be used to manage rpm/deb packages like [Apt]<https://en.wikipedia.org/wiki/APT_(software)>,
or [Yum]<https://www.redhat.com/sysadmin/how-manage-packages>

# Setup customized repo
You just need to do the setup of intelchain repo once on a new host.
**TODO**: the repo in this document are for development/testing purpose only.

Official production repo will be different.

## RPM Package
RPM is for Redhat/Fedora based Linux distributions, such as Amazon Linux and CentOS.

```bash
# do the following once to add the intelchain development repo
curl -LsSf http://haochen-intelchain-pub.s3.amazonaws.com/pub/yum/intelchain-dev.repo | sudo tee -a /etc/yum.repos.d/intelchain-dev.repo
sudo rpm --import https://raw.githubusercontent.com/zennittians/intelchain-open/master/intelchain-release/intelchain-pub.key
```

## Deb Package
Deb is supported on Debian based Linux distributions, such as Ubuntu, MX Linux.

```bash
# do the following once to add the intelchain development repo
curl -LsSf https://raw.githubusercontent.com/zennittians/intelchain-open/master/intelchain-release/intelchain-pub.key | sudo apt-key add
echo "deb http://haochen-intelchain-pub.s3.amazonaws.com/pub/repo bionic main" | sudo tee -a /etc/apt/sources.list

```

# Test cases
## installation
```
# debian/ubuntu
sudo apt-get update
sudo apt-get install intelchain

# fedora/amazon linux
sudo yum install intelchain
```
## configure/start
```
# dpkg-reconfigure intelchain (TODO)
sudo systemctl start intelchain
```

## uninstall
```
# debian/ubuntu
sudo apt-get remove intelchain

# fedora/amazon linux
sudo yum remove intelchain
```

## upgrade
```bash
# debian/ubuntu
sudo apt-get update
sudo apt-get upgrade

# fedora/amazon linux
sudo yum update --refresh
```

## reinstall
```bash
remove and install
```

# Rclone
## install latest rclone
```bash
# debian/ubuntu
curl -LO https://downloads.rclone.org/v1.52.3/rclone-v1.52.3-linux-amd64.deb
sudo dpkg -i rclone-v1.52.3-linux-amd64.deb

# fedora/amazon linux
curl -LO https://downloads.rclone.org/v1.52.3/rclone-v1.52.3-linux-amd64.rpm
sudo rpm -ivh rclone-v1.52.3-linux-amd64.rpm
```

## do rclone
```bash
# validator runs on shard1
sudo -u intelchain intelchain-rclone.sh /home/intelchain 0
sudo -u intelchain intelchain-rclone.sh /home/intelchain 1

# explorer node
sudo -u intelchain intelchain-rclone.sh -a /home/intelchain 0
```

# Setup explorer (non-validating) node
To setup an explorer node (non-validating) node, please run the `intelchain-setup.sh` at first.

```bash
sudo /usr/sbin/intelchain-setup.sh -t explorer -s 0
```
to setup the node as an explorer node w/o blskey setup.

# Setup new validator
Please copy your blskey to `/home/intelchain/.itc/blskeys` directory, and start the node.
The default configuration is for validators on mainnet. No need to run `intelchain-setup.sh` script.

# Start/stop node
* `systemctl start intelchain` to start node
* `systemctl stop intelchain` to stop node
* `systemctl status intelchain` to check status of node

# Change node configuration
The node configuration file is in `/etc/intelchain/intelchain.conf`.  Please edit the file as you need.
```bash
sudo vim /etc/intelchain/intelchain.conf
```

# Support
Please open new github issues in https://github.com/zennittians/intelchain/issues.
