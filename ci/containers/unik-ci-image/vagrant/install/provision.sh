#!/usr/bin/env bash

set -e

curl -sSL https://get.docker.com/ | sh
sudo usermod -aG docker $USER
sudo add-apt-repository ppa:ubuntu-lxc/lxd-stable
sudo apt-get update
sudo apt-get -y git curl jq make golang virtualbox vagrant

vagrant plugin install vagrant-vsphere

echo 'export GOPATH=${HOME}/go' >> ${HOME}/.bashrc
echo 'export PATH=${PATH}:${HOME}/go/bin' >> ${HOME}/.bashrc
source ${HOME}/.bashrc

mkdir -p $GOPATH/src/github.com/emc-advanced-dev/
go get -u github.com/jteeuwen/go-bindata/...

cd $GOPATH/src/github.com/emc-advanced-dev/
git clone https://github.com/emc-advanced-dev/unik


#set up vbox networks
NET=$(VBoxManage list hostonlyifs | grep vboxnet0)
if [ -z "$NET" ]
then
    VBoxManage hostonlyif create
fi

NAT=$(VBoxManage list natnetworks | grep nat)
if [ -z "$NAT" ]
then
    VBoxManage natnetwork add --netname nat --network eth0
fi

#TODO: start natnetwork