#!/usr/bin/env bash
set -e
set -x
PROJECT_ROOT="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )/../../" && \
PROJECT_ROOT="$(cd ${PROJECT_ROOT} && pwd)"
echo "testing virtualbox" && \
echo "project root is ${PROJECT_ROOT}" && \
    MAKE_CONTAINERS=0 \
    TEST_VIRTUALBOX=1 \
    VBOX_ADAPTER_NAME=vboxnet1 \
    VBOX_ADAPTER_TYPE=host_only \
    ginkgo -r -v
