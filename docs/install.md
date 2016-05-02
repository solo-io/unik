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
##if installing with vsphere support:
$ VSPHERE=1 make install
##else
$ make install
```
Continue to [configuration](configure.md) to learn how to configure your unik setup

Note: we recommend [removing intermediate docker containers](http://jimhoskins.com/2013/07/27/remove-untagged-docker-images.html) after building is finished. This will free up disk space, at the cost of making re-building containers slower should you find the need to do so again.

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
