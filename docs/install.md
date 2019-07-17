# Installing UniK

#### Prerequisites:
- [Docker](http://www.docker.com/) installed and running (Docker machine and Docker for Mac are supported)
- [`jq`](https://stedolan.github.io/jq/)
- [`make`](https://www.gnu.org/software/make/)
- [Virtualbox](https://www.virtualbox.org/) (if using the [virtualbox provider](providers/virtualbox.md))
- [go-bindata](https://github.com/jteeuwen/go-bindata)

---
#### Install
```
$ git clone https://github.com/solo-io/unik.git
$ cd unik
$ make # or 'make binary'; see the notes below
$ _build/unik
Unik is a tool for compiling application source code
into bootable disk images. Unik also runs and manages unikernel
instances across infrastructures.
...
```

Note that `make` will pre-pull a number of large docker containers unik needs in order to run. In order to skip pre-pulling (as you may not require all of these containers), you can replace `make` with `make binary`. Note that the first time unik requires a specific docker image, it will pull that image.

This will place the `unik` executable at **unik/_build/unik**. Run UniK commands with `./_build/unik`, or move the binary to somewhere in your path, such as `/usr/local/bin` to run commands from anywhere with `unik [command]`

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
