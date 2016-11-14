SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

define pull_container
	docker pull projectunik/$(1):$(shell jq '.["$(1)"]' containers/versions.json)
endef

define update_container_dependency
	cat containers/versions.json | jq .['"$(2)"'] -r
	$(eval BASE_VERSION=$(shell cat containers/versions.json | jq .['"$(2)"'] -r))
	echo $(BASE_VERSION)
	cd containers/$(1) && perl -pi -e 's/FROM projectunik\/(.*):.*/FROM projectunik\/$$1:$(BASE_VERSION)/g' Dockerfile$(3)
endef

define update_version_bindata
	go-bindata -pkg versiondata -o containers/container-versions.go containers/versions.json
endef

define update_container_version
	echo $(2) > tmpfile
	cat containers/versions.json | jq .['"$(1)"']=\"`cat tmpfile`\" > containers/versions.json
	rm tmpfile
	$(call update_version_bindata)
endef

define build_container
	$(eval BASE_CONTAINER=$(shell cd containers/$(1) && cat Dockerfile$(3) | grep FROM | perl -p -e 's/FROM projectunik\/(.*):.*/$$1/g'))
	echo $(BASE_CONTAINER)
	$(if $(findstring FROM,$(BASE_CONTAINER)),,$(call update_container_dependency,$(1),$(BASE_CONTAINER),$(3)))
	cd containers/$(1) && docker build -t projectunik/$(2):build -f Dockerfile$(3) .
	$(eval CONTAINER_TAG=$(shell echo 'docker inspect projectunik/$(2):build'))
	$(eval CONTAINER_TAG=$(shell echo '$(CONTAINER_TAG) | jq .[].Id' -r ))
	$(eval CONTAINER_TAG=$(shell echo '$(CONTAINER_TAG) | sed 's/sha256://g'' ))
	$(eval CONTAINER_TAG=$(shell echo '$(CONTAINER_TAG) | head -c 16' ))
	$(eval CONTAINER_TAG=$(shell echo '$$$$($(CONTAINER_TAG))' ))
	docker tag projectunik/$(2):build projectunik/$(2):$(CONTAINER_TAG)
	$(call update_container_version,$(2),$(CONTAINER_TAG))
	@echo Built projectunik/$(2):$(CONTAINER_TAG)
	docker rmi projectunik/$(2):build
endef

define remove_container
	docker rmi -f projectunik/$(1):$(shell cat containers/versions.json  | jq '.["$(1)"]')
endef

all: ${SOURCES} binary

.PHONY: pull
.PHONY: containers
.PHONY: rump-debugger-qemu
.PHONY: compilers-includeos-cpp-common
.PHONY: compilers-includeos-cpp-hw
.PHONY: compilers-rump-base-common
.PHONY: compilers-rump-base-hw
.PHONY: compilers-rump-base-xen
.PHONY: compilers-rump-java-hw
.PHONY: compilers-rump-go-hw
.PHONY: compilers-rump-go-xen
.PHONY: compilers-rump-nodejs-hw
.PHONY: compilers-rump-nodejs-hw-no-stub
.PHONY: compilers-rump-nodejs-xen
.PHONY: compilers-rump-c-hw
.PHONY: compilers-rump-c-xen
.PHONY: compilers-rump-python3-hw
.PHONY: compilers-rump-python3-hw-no-stub
.PHONY: compilers-rump-python3-xen
.PHONY: compilers-osv-java
.PHONY: compilers-mirage-ocaml-xen
.PHONY: compilers-mirage-ocaml-ukvm

.PHONY: compilers
.PHONY: boot-creator
.PHONY: image-creator
.PHONY: vsphere-client
.PHONY: qemu-util
.PHONY: utils

#pull containers
pull:
	@echo "Pullling containers from docker hub"
	$(call pull_container,vsphere-client)
	$(call pull_container,boot-creator)
	$(call pull_container,qemu-util)
	$(call pull_container,compilers-includeos-cpp-hw)
	$(call pull_container,compilers-osv-java)
	$(call pull_container,compilers-rump-java-hw)
	$(call pull_container,compilers-rump-java-xen)
	$(call pull_container,compilers-rump-go-hw)
	$(call pull_container,compilers-rump-go-xen)
	$(call pull_container,compilers-rump-nodejs-hw)
	$(call pull_container,compilers-rump-nodejs-hw-no-stub)
	$(call pull_container,compilers-rump-nodejs-xen)
	$(call pull_container,compilers-rump-c-hw)
	$(call pull_container,compilers-rump-c-xen)
	$(call pull_container,compilers-rump-python3-hw)
	$(call pull_container,compilers-rump-python3-hw-no-stub)
	$(call pull_container,compilers-rump-python3-xen)
	$(call pull_container,compilers-rump-base-xen)
	$(call pull_container,compilers-rump-base-hw)
	$(call pull_container,rump-debugger-qemu)
	$(call pull_container,compilers-rump-base-common)
	docker pull euranova/ubuntu-vbox
#------

#build containers from source
containers: compilers utils
	echo "Built containers from source"

#compilers
compilers: compilers-includeos-cpp-hw \
           compilers-rump-java-hw \
           compilers-rump-java-xen \
           compilers-rump-go-hw \
           compilers-rump-go-xen \
           compilers-rump-nodejs-hw \
           compilers-rump-nodejs-hw-no-stub \
           compilers-rump-nodejs-xen \
           compilers-rump-c-hw \
           compilers-rump-c-xen \
           compilers-rump-python3-hw \
           compilers-rump-python3-hw-no-stub \
           compilers-rump-python3-xen \
           compilers-osv-java

compilers-includeos-cpp-common:
	$(call build_container,compilers/includeos/cpp,$@,.common)

compilers-includeos-cpp-hw: compilers-includeos-cpp-common
	$(call build_container,compilers/includeos/cpp,$@,.hw)

compilers-rump-base-common:
	$(call build_container,compilers/rump/base,$@,.common)

compilers-rump-base-hw: compilers-rump-base-common
	$(call build_container,compilers/rump/base,$@,.hw)

compilers-rump-base-xen: compilers-rump-base-common
	$(call build_container,compilers/rump/base,$@,.xen)

compilers-rump-java-hw: compilers-rump-base-hw
	cd containers/compilers/rump/java/java-wrapper && mvn package
	$(call build_container,compilers/rump/java,$@,.hw)

compilers-rump-java-xen: compilers-rump-base-xen
	cd containers/compilers/rump/java/java-wrapper && mvn package
	$(call build_container,compilers/rump/java,$@,.xen)

rump-debugger-qemu: compilers-rump-base-hw
	$(call build_container,debuggers/rump/base,$@,.hw)

rump-debugger-xen: compilers-rump-base-xen
	$(call build_container,debuggers/rump/base,$@,.xen)

compilers-rump-go-hw: compilers-rump-base-hw
	$(call build_container,compilers/rump/go,$@,.hw)

compilers-rump-go-xen: compilers-rump-base-xen
	$(call build_container,compilers/rump/go,$@,.xen)

compilers-rump-nodejs-hw: compilers-rump-base-hw
	$(call build_container,compilers/rump/nodejs,$@,.hw)

compilers-rump-nodejs-hw-no-stub: compilers-rump-base-hw
	$(call build_container,compilers/rump/nodejs,$@,.hw.no-stub)

compilers-rump-nodejs-xen: compilers-rump-base-xen
	$(call build_container,compilers/rump/nodejs,$@,.xen)

compilers-rump-c-hw: compilers-rump-go-hw
	$(call build_container,compilers/rump/c,$@,.hw)

compilers-rump-c-xen: compilers-rump-go-xen
	$(call build_container,compilers/rump/c,$@,.xen)

compilers-rump-python3-hw: compilers-rump-go-hw
	$(call build_container,compilers/rump/python3,$@,.hw)

compilers-rump-python3-hw-no-stub: compilers-rump-base-hw
	$(call build_container,compilers/rump/python3,$@,.hw.no-stub)

compilers-rump-python3-xen: compilers-rump-go-xen
	$(call build_container,compilers/rump/python3,$@,.xen)

compilers-osv-java:
	cd containers/compilers/osv/java-compiler/java-main-caller && mvn package
	cd containers/compilers/osv/java-compiler && GOOS=linux go build
	$(call build_container,compilers/osv/java-compiler,$@,)
	cd containers/compilers/osv/java-compiler && rm java-compiler
	cd containers/compilers/osv/java-compiler/java-main-caller && rm -rf target

compilers-mirage-ocaml-xen:
	$(call build_container,compilers/mirage/ocaml,$@,.xen)

compilers-mirage-ocaml-ukvm:
	$(call build_container,compilers/mirage/ocaml,$@,.ukvm)

#utils
utils: boot-creator image-creator vsphere-client qemu-util

boot-creator:
	cd containers/utils/boot-creator && GO15VENDOREXPERIMENT=1 GOOS=linux go build
	$(call build_container,utils/boot-creator,$@,)
	cd containers/utils/boot-creator && rm boot-creator

image-creator:
	cd containers/utils/image-creator && GO15VENDOREXPERIMENT=1 GOOS=linux go build
	$(call build_container,utils/image-creator,$@,)
	cd containers/utils/image-creator && rm image-creator

vsphere-client: containers/utils/vsphere-client/vsphere-client.empty

VSPHERE_CLIENT_SOURCES := containers/utils/vsphere-client/Dockerfile containers/utils/vsphere-client/pom.xml  $(shell find containers/utils/vsphere-client/src/)
containers/utils/vsphere-client/vsphere-client.empty: $(VSPHERE_CLIENT_SOURCES)
	cd containers/utils/vsphere-client && mvn package
	$(call build_container,utils/vsphere-client,vsphere-client,)
	cd containers/utils/vsphere-client && rm -rf target
	touch containers/utils/vsphere-client/vsphere-client.empty

qemu-util:
	$(call build_container,utils/qemu-util,$@,)

#------

#binary

BINARY=unik

# don't override if provided already
ifeq (,$(TARGET_OS))
	UNAME:=$(shell uname)
	ifeq ($(UNAME),Linux)
		TARGET_OS:=linux
	else ifeq ($(UNAME),Darwin)
		TARGET_OS:=darwin
	endif
endif

binary: ${SOURCES}
ifeq (,$(TARGET_OS))
	echo "Unknown platform $(UNAME)"
	echo "Unknown platform $(TARGET_OS)"
	exit 1
endif
	@echo Building for platform $(UNAME)
	docker build -t projectunik/$@ -f Dockerfile .
	mkdir -p ./_build
	docker run --rm -v $(shell pwd)/_build:/opt/build -e TARGET_OS=$(TARGET_OS) projectunik/$@
	#docker rmi -f projectunik/$@
	@echo "Install finished! UniK binary can be found at $(shell pwd)/_build/unik"
#----

# local build - useful if you have development env setup. if not - use binary! (this can't depend on binary as binary depends on it via the Dockerfile)
localbuild: instance-listener/bindata/instance_listener_data.go containers/version-data.go ${SOURCES}
	GOOS=${TARGET_OS} go build -v .

# local install - useful if you have development env setup. if not - use binary! (this can't depend on binary as binary depends on it via the Dockerfile)
localinstall: instance-listener/bindata/instance_listener_data.go containers/version-data.go ${SOURCES}
	GOOS=${TARGET_OS} go install -v .

containers/version-data.go: containers/versions.json
	$(call update_version_bindata)

instance-listener/bindata/instance_listener_data.go: instance-listener/main.go
	go-bindata -pkg bindata -o instance-listener/bindata/instance_listener_data.go --ignore=instance-listener/bindata/ instance-listener/...

#clean up
.PHONY: uninstall remove-containers clean

uninstall:
	rm $(which ${BINARY})

remove-containers:
	-docker rmi -f projectunik/binary
	-$(call remove_container,vsphere-client)
	-rm -rf containers/utils/vsphere-client/vsphere-client.empty
	-$(call remove_container,image-creator)
	-$(call remove_container,boot-creator)
	-$(call remove_container,compilers-osv-java)
	-$(call remove_container,compilers-rump-go-xen)
	-$(call remove_container,compilers-rump-go-hw)
	-$(call remove_container,compilers-rump-nodejs-hw)
	-$(call remove_container,compilers-rump-nodejs-hw-no-stub)
	-$(call remove_container,compilers-rump-nodejs-xen)
	-$(call remove_container,compilers-rump-c-hw)
	-$(call remove_container,compilers-rump-c-xen)
	-$(call remove_container,compilers-rump-python3-hw)
	-$(call remove_container,compilers-rump-python3-hw-no-stub)
	-$(call remove_container,compilers-rump-python3-xen)
	-$(call remove_container,compilers-rump-base-xen)
	-$(call remove_container,compilers-rump-base-hw)
	-$(call remove_container,rump-debugger-qemu)
	-$(call remove_container,compilers-rump-base-common)

clean:
	rm -rf ./_build
#---
