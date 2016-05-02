<img src="http://i.imgur.com/npkzp8l.png" alt="Build and run unikernels" width="141" height="50">

UniK (pronounced you-neek) is a tool for compiling application sources into unikernels (lightweight bootable disk images) rather than binaries. UniK runs and manages instances of compiled images across a variety of cloud providers as well as locally on Virtualbox. UniK utilizes a simple docker-like command line interface, making building unikernels as easy as building containers. 

UniK is built to be easily extensible, allowing (and encouraging) adding support for unikernel [compilers](docs/compilers/README.md) and cloud  [providers](docs/providers/README.md). See [architecture](docs/architecture.md) for a better understanding of UniK's pluggable code design.

---
### Documentation
- **Installation**
  - [Install UniK](docs/install.md)
  - [Configuring the daemon](docs/configure.md)
- **Getting Started**
  - [Run your first unikernel](docs/getting_started.md) with UniK
- **User Documenation**
  - [Provider configuration](docs/config.md)
  - Using the [cli](docs/cli.md)
- **Developer Documentation**
  - UniK's [REST API](docs/api.md)
  - Adding [compiler](docs/compilers/README.md) support
  - Adding [provider](docs/providers/README.md) support

---
### Supported unikernel types:
* **(go)rump**: UniK supports compiling Go code into [rumpkernels](docs/compilers/rump.md)

*We are looking for community help to add support for more unikernel types and languages.*

### Supported providers:
* [Virtualbox](docs/providers/virtualbox.md)
* [AWS](docs/providers/aws.md)
* [vSphere](docs/providers/vsphere.md)

### Roadmap:
* c++ and nodejs support using [rump kernel](http://rumpkernel.org)
* java support using [OSv](http://osv.io/)
* ocaml support using [MirageOs](https://mirage.io/)
* additional provider support including [OpenStack](https://www.openstack.org/)
* dynamic volume and parameter configuration at instance runtime (rather than at compile time)
* adding a test suite
* better code documentation
* `unik pull` & `unik push` && unikhub for sharing unikernel images
* multi-account support per provider (i.e. multiple AWS accounts/regions, etc.)

### Known bugs:
* **time.Sleep() in (go)rump**: time.Sleep currently causes panic() in gorump unikernels
