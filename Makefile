all: pull ${SOURCES}

#pull containers
.PHONY: pull:
	echo "Pullling containers from docker hub"
	docker pull projectunik/vsphere-client
	docker pull projectunik/image-creator
	docker pull projectunik/boot-creator
	docker pull projectunik/compilers-rump-go-xen
	docker pull projectunik/compilers-rump-go-hw
	docker pull projectunik/compilers-rump-base-xen
	docker pull projectunik/compilers-rump-base-hw
	docker pull projectunik/compilers-rump-base-common
#------

#build containers from source
.PHONY: containers: compilers utils
	echo "Built containers from source"

#compilers
.PHONY: compilers-rump-base-common:
	cd containers/compilers/rump/base && docker build -t projectunik/$@ -f Dockerfile.common .

.PHONY: compilers-rump-base-hw: compilers-rump-base-common
	cd containers/compilers/rump/base && docker build -t projectunik/$@ -f Dockerfile.hw .

.PHONY: compilers-rump-base-xen: compilers-rump-base-common
	cd containers/compilers/rump/base && docker build -t projectunik/$@ -f Dockerfile.xen .

.PHONY: compilers-rump-go-hw: compilers-rump-base-hw
	cd containers/compilers/rump/go && docker build -t projectunik/$@ -f Dockerfile.hw .

.PHONY: compilers-rump-go-xen: compilers-rump-base-xen
	cd containers/compilers/rump/go && docker build -t projectunik/$@ -f Dockerfile.xen .

.PHONY: compilers: compilers-rump-go-hw compilers-rump-go-xen

#utils
.PHONY: boot-creator:
	cd containers/utils/boot-creator && GOOS=linux go build && docker build -t projectunik/$@ -f Dockerfile . && rm boot-creator

.PHONY: image-creator:
	cd containers/utils/image-creator && GOOS=linux go build && docker build -t projectunik/$@ -f Dockerfile . && rm image-creator

.PHONY: vsphere-client:
	cd containers/utils/vsphere-client && mvn package && docker build -t projectunik/$@ -f Dockerfile . && rm -rf target

.PHONY: utils: boot-creator image-creator vsphere-client
#------

#binary & install
SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=unik

install: all ${SOURCES}
	go install

#----

#clean up
.PHONY: uninstall
	rm $(which ${BINARY})

.PHONY: remove-containers
	-docker rmi -f projectunik/vsphere-client
	-docker rmi -f projectunik/image-creator
	-docker rmi -f projectunik/boot-creator
	-docker rmi -f projectunik/compilers-rump-go-xen
	-docker rmi -f projectunik/compilers-rump-go-hw
	-docker rmi -f projectunik/compilers-rump-base-xen
	-docker rmi -f projectunik/compilers-rump-base-hw
	-docker rmi -f projectunik/compilers-rump-base-common

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi
#---
