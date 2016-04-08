
define buildcontainer
$(eval words = $(subst -, ,$@))
$(eval folder = $(word 1, $(words)))
$(eval framework = $(word 2, $(words)))
$(eval platform = $(word 3, $(words)))
docker build -t $@ -f $(folder)/$(framework)/Dockerfile.$(platform) $(folder)/$(framework)
endef

define buildsimplecontainer
docker build -t $@ $@
endef


.PHONY: all build rumpstager rumpcompiler-base-common rumpcompiler-base-hw rumpcompiler-base-xen rumpcompiler-go-hw rumpcompiler-go-xen rumpcdebugger-gdb-hw vsphere-client osvcompiler osvec2stager

all: build

rumpcompiler-base-common:
	$(buildcontainer)

rumpcompiler-base-hw: rumpcompiler-base-common
	$(buildcontainer)

rumpcompiler-base-xen: rumpcompiler-base-common
	$(buildcontainer)

rumpcompiler-go-hw: rumpcompiler-base-hw
	$(buildcontainer)

rumpcompiler-go-xen: rumpcompiler-base-xen
	$(buildcontainer)

rumpcompiler-pi-hw: rumpcompiler-base-hw
	$(buildcontainer)

rumpdebugger-gdb-hw: rumpcompiler-base-xen
	$(buildcontainer)

rumpdebugger-gdb-arm: rumpcompiler-pi-hw
	$(buildcontainer)

rumpstager:
	$(buildsimplecontainer)

vsphere-client:
	$(buildsimplecontainer)

osvcompiler:
	$(buildsimplecontainer)

osvec2stager:
	$(buildsimplecontainer)

build: rumpstager rumpcompiler-base-hw rumpcompiler-base-xen rumpcompiler-go-hw rumpcompiler-go-xen rumpcdebugger-gdb-hw vsphere-client osvcompiler osvec2stager
