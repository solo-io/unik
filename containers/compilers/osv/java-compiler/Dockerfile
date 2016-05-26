FROM ubuntu:16.04

ENV GOPATH=$HOME/go
ENV GOBIN=$GOPATH/bin
ENV PATH=$GOBIN:/usr/local/go/bin:$PATH
RUN apt-get update -y
RUN apt-get install -y qemu maven wget git openjdk-7-jdk curl
RUN curl https://storage.googleapis.com/golang/go1.5.2.linux-amd64.tar.gz | tar xz -C /usr/local && mv /usr/local/go /usr/local/go1.5 && ln -s /usr/local/go1.5 /usr/local/go
RUN mkdir -p $GOPATH/src/github.com/cloudius-systems
RUN cd $GOPATH/src/github.com/cloudius-systems && git clone https://github.com/emc-advanced-dev/capstan
RUN cd $GOPATH/src/github.com/cloudius-systems/capstan && ./install
RUN apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

VOLUME /project_directory

#Pre-pull osv base image
RUN capstan pull cloudius/osv-openjdk

COPY java-main-caller-udp-bootstrap/ /java-main-caller-udp-bootstrap/
COPY java-main-caller-ec2-bootstrap/ /java-main-caller-ec2-bootstrap/
COPY java-compiler /

ENTRYPOINT ["/java-compiler"]

#run this container with
#docker run --rm --privileged -v SOURCE_ROOT:/project_directory projectunik/osv-java-compiler
