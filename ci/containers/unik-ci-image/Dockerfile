FROM ubuntu:16.04

RUN DEBIAN_FRONTEND=noninteractive apt-get update -y && \
    apt-get install -y curl

RUN curl -O https://releases.hashicorp.com/vagrant/1.8.5/vagrant_1.8.5_x86_64.deb && \
    dpkg -i vagrant_1.8.5_x86_64.deb && \
    apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN vagrant plugin install vagrant-vsphere
RUN mkdir /vagrant
COPY /vagrant /vagrant/
RUN cd /vagrant && vagrant box add vsphere-unik-test-box ./vsphere-unik-test.box