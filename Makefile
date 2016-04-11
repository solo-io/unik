define buildcontainer
$(eval words = $(subst -, ,$@))
$(eval type      = $(word 1, $(words)))
$(eval framework = $(word 2, $(words)))
$(eval proglang  = $(word 3, $(words)))
$(eval platform  = $(word 4, $(words)))

$(eval ifneq ($(platform),)
    platform = .$(platform) 
 endif)


cd containers && docker build -t $@ -f $(type)/$(framework)/$(proglang)/Dockerfile$(platform) $(type)/$(framework)/$(proglang)
endef

define cmdbuildcontainer
$(eval words = $(subst /, ,$@))
$(eval folder = $(word 1, $(words)))
$(eval name = $(word 2, $(words)))
docker build -t $(name) $@
endef

define gobuild
cd $@ && go build
endef

define gobuild-linux
cd $@ && GOOS=linux go build
endef


SUBDIRS = cmd/boot-creator cmd/image-creator cmd/stager cmd/provider cmd/volume-uploader cmd/daemon

all: $(SUBDIRS)
.PHONY: all $(SUBDIRS)

cmd/daemon:
	$(gobuild)
	$(cmdbuildcontainer)

cmd/stager:
	$(gobuild)

cmd/provider:
	$(gobuild)

cmd/image-creator:
	$(gobuild-linux)
	$(cmdbuildcontainer)

cmd/boot-creator:
	$(gobuild-linux)
	$(cmdbuildcontainer)

cmd/volume-uploader:
	$(gobuild)

# rump compilers - these produce a .bin unikernel

compilers-rump-base-common:
	$(buildcontainer)

compilers-rump-base-hw: compilers-rump-base-common
	$(buildcontainer)

compilers-rump-base-xen: compilers-rump-base-common
	$(buildcontainer)

compilers-rump-go-hw: compilers-rump-base-hw
	$(buildcontainer)

compilers-rump-go-xen: compilers-rump-base-xen
	$(buildcontainer)

compilers-mirage-ocaml:
	$(buildcontainer)


