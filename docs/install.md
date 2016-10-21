# Installing UniK

#### Prerequisites:
- [Docker](http://www.docker.com/) installed and running with at least 8GB available space for building images. **Note: *Docker for Mac* will not work as it does not support Linux's device mapper. Until we find an alternative for building disk images, if you are running Unik on OS X, you will need to use `docker-machine` (part of the Docker Toolbox for Mac).**
- [`jq`](https://stedolan.github.io/jq/)
- [`make`](https://www.gnu.org/software/make/)
- [Virtualbox](https://www.virtualbox.org/) (if using the [virtualbox provider](providers/virtualbox.md))

---
#### Install
```
$ git clone https://github.com/emc-advanced-dev/unik.git
$ cd unik
$ make
Unik is a tool for compiling application source code
into bootable disk images. Unik also runs and manages unikernel
instances across infrastructures.
...
```

This will place the `unik` executable at **$GOPATH/bin/unik**. Run UniK commands with `$GOPATH/bin/unik`, or add `$GOPATH/bin` to your `$PATH`.

Continue to [configuration](configure.md) to learn how to configure your UniK setup
---
#### Building Containers from Source
By default, `make` will pull all of the necessary container images from [Docker Hub](https://hub.docker.com/).
If you wish to build containers from sources, you will need:
- [Golang](https://golang.org/) v1.5 or later
- `$GOPATH` should be set and `$GOPATH/bin` should be part of your `$PATH` (see https://golang.org/doc/code.html#GOPATH)
- [Apache Maven](https://maven.apache.org/)
Verify that `mvn` and `go` are installed, and your `$GOPATH` is set correctly. Then simply:

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
