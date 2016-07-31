FROM ubuntu:16.04

RUN apt-get update -y && \
 apt-get install -y software-properties-common && \
 add-apt-repository -y ppa:openjdk-r/ppa && \
 apt-get update -y && \
 apt-get install -y qemu maven wget git openjdk-7-jdk curl && \
 apt-get install -y build-essential && \
 apt-get clean -y && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*


RUN curl https://storage.googleapis.com/golang/go1.5.2.linux-amd64.tar.gz | tar xz -C /usr/local && mv /usr/local/go /usr/local/go1.5 && ln -s /usr/local/go1.5 /usr/local/go


ENV GOPATH=/go
ENV GOBIN=$GOPATH/bin
ENV PATH=$GOBIN:/usr/local/go/bin:$PATH

RUN mkdir -p $GOPATH/src/github.com/cloudius-systems
RUN cd $GOPATH/src/github.com/cloudius-systems && git clone https://github.com/emc-advanced-dev/capstan
RUN cd $GOPATH/src/github.com/cloudius-systems/capstan && ./install

RUN capstan pull cloudius/osv-openjdk8

VOLUME /project_directory
COPY java-main-caller/target/jar-wrapper-1.0-SNAPSHOT-jar-with-dependencies.jar /program.jar

#Build base jar runner
COPY Capstanfile-jar /tmp/Capstanfile-jar

RUN mkdir /jar-runner/ && \
    mv /tmp/Capstanfile-jar /jar-runner/Capstanfile && \
    cd /jar-runner/ && \
    capstan build unik-jar-runner &&\
    rm -rf /jar-runner

#Build base tomcat image
COPY Capstanfile-war /tmp/Capstanfile-war

RUN cd / && git clone https://github.com/cloudius-systems/osv-apps && \
    cd /osv-apps/tomcat && \
    make && \
    cd /osv-apps/tomcat && sed -i -e 's/port="8081"/port="${port.http.nonssl}"/g' ROOTFS/usr/tomcat/conf/server.xml && \
    mv /tmp/Capstanfile-war /osv-apps/tomcat/Capstanfile && \
    cd /osv-apps/tomcat && \
    capstan build unik-tomcat &&\
    rm -rf /osv-apps

COPY java-compiler /

ENTRYPOINT ["/java-compiler"]

#run this container with
#docker run --rm --privileged -v SOURCE_ROOT:/project_directory projectunik/osv-java-compiler
