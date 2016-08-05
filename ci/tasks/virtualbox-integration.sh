#!/usr/bin/env bash

set -e

cd vagrant-out
vagrant ssh -c 'cd ${HOME}/go/src/github.com/emc-advanced-dev/ && GOPATH=${HOME}/go PATH=${PATH}:${HOME}/go/bin TEST_VIRTUALBOX=1 bash test/scripts/test_ginkgo.sh -failFast'
