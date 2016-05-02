# Installing UniK


#### Prerequisites:
- [Docker](http://www.docker.com/) installed and running with at least 8GB available space for building images
- [`make`](https://www.gnu.org/software/make/)
- [Golang](https://golang.org/) v1.5 or later
- `$GOPATH` should be set and `$GOPATH/bin` should be part of your `$PATH` (see https://golang.org/doc/code.html#GOPATH)
- [Virtualbox](https://www.virtualbox.org/) (if using the [virtualbox provider](providers/virtualbox.md))
- [Apache Maven](https://maven.apache.org/) (if using the [vsphere provider](providers/vsphere.md))

---
#### Install
```
$ mkdir -p $GOPATH/src/github.com/emc-advanced-dev
$ cd $GOPATH/src/github.com/emc-advanced-dev
$ git clone https://github.com/emc-advanced-dev/unik.git
$ cd unik
$ make install
```
Continue to [configuration](configure.md) to learn how to configure your unik setup

---
#### Building Containers from Source
By default, `make install` will pull all of the necessary container images from [Docker Hub](https://hub.docker.com/). If you wish to build containers from sources, simply run
```
$ make containers
```
---
#### Uninstall

##### `unik` binary
```
$ make uninstall
```

##### UniK docker containers
```
$ make remove-containers
```
