#!/usr/bin/env bash
set -e
set -x
echo Running AWS test
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../../" && \
PROJECT_ROOT="$(cd ${PROJECT_ROOT} && pwd)"
echo "project root is ${PROJECT_ROOT}" && \
#    TEST_VIRTUALBOX=1 \
#    VBOX_ADAPTER_NAME=vboxnet14 \
#    VBOX_ADAPTER_TYPE=host_only \
    TEST_AWS=1 \
    AWS_REGION=us-west-1 \
    AWS_AVAILABILITY_ZONE=us-west-1b \
    ginkgo -r -v