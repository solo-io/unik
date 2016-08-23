#!/usr/bin/env bash

set -x -e
cp /opt/rumprun-packages/openjdk8/bin/java.bin /opt/code/program.bin
cp -r /opt/rumprun-packages/openjdk8/build/javadist/jvm/openjdk-1.8.0-internal/ /opt/code/jdk
cp /tmp/build/program.jar /opt/code/
if [[ "$MAIN_FILE" == *.war ]]; then
    echo "building jetty unikernel"
    cp -r /tmp/build/jetty-distribution-*/ /opt/code/jetty
    mv /opt/code/$MAIN_FILE /opt/code/jetty/webapps/$MAIN_FILE
fi
