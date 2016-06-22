FROM ubuntu:14.04

ENV GOPATH=$HOME/go
ENV GOBIN=$GOPATH/bin
ENV PATH=$GOBIN:/usr/local/go/bin:$PATH
RUN apt-get update -y
RUN apt-get install -y qemu maven wget git openjdk-7-jdk curl
RUN apt-get install -y build-essential
RUN curl https://storage.googleapis.com/golang/go1.5.2.linux-amd64.tar.gz | tar xz -C /usr/local && mv /usr/local/go /usr/local/go1.5 && ln -s /usr/local/go1.5 /usr/local/go
RUN mkdir -p $GOPATH/src/github.com/cloudius-systems
RUN cd $GOPATH/src/github.com/cloudius-systems && git clone https://github.com/emc-advanced-dev/capstan
RUN cd $GOPATH/src/github.com/cloudius-systems/capstan && ./install
RUN apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN capstan pull cloudius/osv-openjdk8

VOLUME /project_directory
COPY java-main-caller/target/jar-wrapper-1.0-SNAPSHOT-jar-with-dependencies.jar /program.jar

#Build base jar runner
RUN mkdir /jar-runner/
COPY Capstanfile-jar /jar-runner/Capstanfile
RUN cd /jar-runner/ && \
    capstan build unik-jar-runner
RUN rm -rf /jar-runner

#Build base tomcat image
RUN cd / && git clone https://github.com/cloudius-systems/osv-apps
RUN cd /osv-apps/tomcat && \
    make
COPY Capstanfile-war /osv-apps/tomcat/Capstanfile
RUN cd /osv-apps/tomcat && \
    capstan build unik-tomcat
RUN rm -rf /osv-apps

COPY java-compiler /

ENTRYPOINT ["/java-compiler"]

#run this container with
#docker run --rm --privileged -v SOURCE_ROOT:/project_directory projectunik/osv-java-compiler
