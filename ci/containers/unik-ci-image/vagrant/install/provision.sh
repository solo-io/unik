#!/usr/bin/env bash

set -e
set -x

curl -sSL https://get.docker.com/ | sh
sudo usermod -aG docker $USER
sudo add-apt-repository ppa:ubuntu-lxc/lxd-stable
sudo apt-get update
sudo apt-get install -y git curl jq make golang virtualbox

echo 'export GOPATH=${HOME}/go' >> ${HOME}/.bashrc
echo 'export PATH=${PATH}:${HOME}/go/bin' >> ${HOME}/.bashrc
source ${HOME}/.bashrc

export GOPATH=${HOME}/go
export PATH=${PATH}:${HOME}/go/bin

mkdir -p $GOPATH/src/github.com/emc-advanced-dev/
go get -u github.com/jteeuwen/go-bindata/...
go get -u github.com/onsi/ginkgo/ginkgo
go get -u github.com/onsi/gomega

cd $GOPATH/src/github.com/emc-advanced-dev/
git clone https://github.com/emc-advanced-dev/unik
# TODO: make containers. will make a much longer build, but iwll test changes to code that affects containers
# cd unik && make containers && make localbuild

set +e
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

set -e

#TODO: start natnetwork