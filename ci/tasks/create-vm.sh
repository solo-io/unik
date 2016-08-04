#!/usr/bin/env bash

set -e

: ${VSPHERE_HOST:?}
: ${COMPUTE_RESOURCE_HOST:?}
: ${RESOURCE_POOL_NAME:?}
: ${VSPHERE_TEMPLATE_NAME:?}
: ${VM_NAME:?}
: ${VSPHERE_USER:?}
: ${VSPHERE_PASSWORD:?}

export VM_NAME=${VM_NAME}${RANDOM}

echo using ${VM_NAME}

cd /vagrant && vagrant up