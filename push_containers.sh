#/bin/bash
for i in boot-creator
compilers-osv-java
compilers-rump-base-common
compilers-rump-base-hw
compilers-rump-base-xen
compilers-rump-go-hw
compilers-rump-go-hw-no-stub
compilers-rump-go-xen
compilers-rump-nodejs-hw
compilers-rump-nodejs-hw-no-stub
compilers-rump-nodejs-xen
compilers-rump-python3-hw
compilers-rump-python3-hw-no-stub
compilers-rump-python3-xen
image-creator
qemu-util
rump-debugger-qemu
vsphere-client; do echo "pushing projectunik/ubuntu:null" && docker push projectunik/ubuntu:null; done
