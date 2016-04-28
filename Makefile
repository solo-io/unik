all: compilers utils unik

.PHONY: compilers utils compilers-rump-base-common compilers-rump-base-hw compilers-rump-base-xen compilers-rump-go-hw compilers-rump-go-xen boot-creator image-creator vsphere-client

compilers-rump-base-common:
	cd containers/compilers/rump/base && docker build -t unik/$@ -f Dockerfile.common .

compilers-rump-base-hw: compilers-rump-base-common
	cd containers/compilers/rump/base && docker build -t unik/$@ -f Dockerfile.hw .

compilers-rump-base-xen: compilers-rump-base-common
	cd containers/compilers/rump/base && docker build -t unik/$@ -f Dockerfile.xen .

compilers-rump-go-hw: compilers-rump-base-hw
	cd containers/compilers/rump/go && docker build -t unik/$@ -f Dockerfile.hw .

compilers-rump-go-xen: compilers-rump-base-xen
	cd containers/compilers/rump/go && docker build -t unik/$@ -f Dockerfile.xen .

compilers: compilers-rump-go-hw compilers-rump-go-xen

boot-creator:
	cd containers/utils/boot-creator && GOOS=linux go build && docker build -t unik/$@ -f Dockerfile . && rm boot-creator

image-creator:
	cd containers/utils/image-creator && GOOS=linux go build && docker build -t unik/$@ -f Dockerfile . && rm image-creator

vsphere-client:
	cd containers/utils/vsphere-client && mvn package && docker build -t unik/$@ -f Dockerfile . && rm -rf target

utils: boot-creator image-creator vsphere-client

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

BINARY=unik

unik: ${SOURCES}
	go build -o ${BINARY} unik.go

.PHONY: install
install: all
	go install unik.go

.PHONY: uninstall
	rm $(which ${BINARY})

.PHONY: clean
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi