<img src="http://i.imgur.com/npkzp8l.png" alt="Build and run unikernels" width="141" height="50">

UniK (pronounced you-neek) is a tool for compiling application sources into unikernels (lightweight bootable disk images) rather than binaries. UniK runs and manages instances of compiled images across a variety of cloud providers as well as locally on Virtualbox. UniK utilizes a simple docker-like command line interface, making building unikernels as easy as building containers.

UniK is built to be easily extensible, allowing (and encouraging) adding support for unikernel [compilers](docs/compilers/README.md) and cloud [providers](docs/providers/README.md). See [architecture](docs/architecture.md) for a better understanding of UniK's pluggable code design.

To learn more about the motivation behind project UniK, read our [blog](https://github.com/emc-advanced-dev/unik/wiki/UniK:-Build-and-Run-Unikernels-with-Ease) post.

---
### Documentation
- **Installation**
  - [Install UniK](docs/install.md)
  - [Configuring the daemon](docs/configure.md)
  - [Launching the InstanceListener](docs/instance_listener.md)
- **Getting Started**
  - [Run your first unikernel](docs/getting_started.md) with UniK
- **User Documenation**
  - Using the [command line interface](docs/cli.md)
  - Compiling [Go](docs/compilers/rump.md#golang) Applications to Unikernels
  - Compiling [Java](docs/compilers/osv.md#java) Applications to Unikernels
  - Compiling [C/C++](docs/compilers/rump.md#c++) Applications to Unikernels
- **Developer Documentation**
  - UniK's [REST API](docs/api.md)
  - Adding [compiler](docs/compilers/README.md) support
  - Adding [provider](docs/providers/README.md) support

---

### Supported unikernel types:
* **rump**: UniK supports compiling C/C++ and Go code into [rumprun](docs/compilers/rump.md) unikernels
* **OSv**: UniK supports compiling Java code into [OSv](http://osv.io/) unikernels (currently only compatible with Virtualbox provider)

*We are looking for community help to add support for more unikernel types and languages.*

### Supported providers:
* [Virtualbox](docs/providers/virtualbox.md)
* [AWS](docs/providers/aws.md)
* [vSphere](docs/providers/vsphere.md)

### Roadmap:
* nodejs support using [rump kernel](http://rumpkernel.org)
* extend [OSv](http://osv.io/) support AWS and vSphere providers
* ocaml support using [MirageOs](https://mirage.io/)
* additional provider support including [OpenStack](https://www.openstack.org/)
* dynamic volume and application arguments configuration at instance runtime (rather than at compile time)
* adding a test suite
* better code documentation
* `unik pull` & `unik push` && unikhub for sharing unikernel images
* multi-account support per provider (i.e. multiple AWS accounts/regions, etc.)
* migrate from [martini](https://github.com/go-martini/martini) to [echo](https://github.com/labstack/echo)
* find an alternative to the [Instance Listener](docs/instance_listener.md) for bootstrapping instances on private networks

UniK is still experimental! APIs and compatibility is subject to change. We are looking for community support to help identify potential bugs and compatibility issues. Please open a Github issue for any problems you may experience, and join us on our [slack channel](http://project-unik.io)

---

### Thanks

**UniK** would not be possible without the valuable open-source work of projects in the unikernel community. We would like to extend a special thank-you to [rumpkernel](https://github.com/rumpkernel/), [deferpanic](https://github.com/deferpanic), and [cloudius systems](https://github.com/cloudius-systems).
