# Getting Started

In this tutorial we'll be:
  1. [installing UniK](getting_started.md#installing-unik)
  2. [writing a simple HTTP Daemon in Go](getting_started.md#write-a-go-http-server)
  3. [compiling to a unikernel and launching an instance on Virtualbox](getting_started.md#compile-an-image-and-run-on-virtualbox)

### Installing UniK
#### Prerequisites
Ensure that each of the following are installed
- [Docker](http://www.docker.com/) installed and running with at least 8GB available space for building images
- [`make`](https://www.gnu.org/software/make/)
- [Golang](https://golang.org/) v1.5 or later
- [Virtualbox](https://www.virtualbox.org/)
- `$GOPATH` should be set and `$GOPATH/bin` should be part of your `$PATH` (see https://golang.org/doc/code.html#GOPATH)

#### Install, configure, and launch UniK
1. Install UniK
  ```
  $ mkdir -p $GOPATH/src/github.com/emc-advanced-dev
  $ cd $GOPATH/src/github.com/emc-advanced-dev
  $ git clone https://github.com/emc-advanced-dev/unik.git
  $ cd unik
  $ make install
  ```
  note: `make install` will take quite a few minutes the first time it runs. the UniK `Makefile` is pulling all of the Docker images that bundle UniK's dependencies.

2. Configure a Host-Only Network on Virtualbox
  * Open Virtualbox
  * Open **Preferences** > **Network** > **Host-only Networks**
  * Click the green add button on the right side of the UI
  * Record the name of the new Host-Only adapter. You will need this in your UniK configuration

3. Configure UniK daemon
  * Using a text editor, create and save the following to `$HOME/.unik/daemon-config.yaml`:
  ```
  providers:
    virtualbox:
      - name: vsphere-1
        adapter_type: host_only
        adapter_name: NEW_HOST_ONLY_ADAPTER
  ```
  replacing `NEW_HOST_ONLY_ADAPTER` with the name of the network adapter you created.

4. Launch UniK and automatically deploy the *Virtualbox Instance Listener*
  * Open a new terminal window/tab. This terminal will be where we leave the UniK daemon running.
  * `cd` to a directory where UniK can download a file.
  * run `unik daemon`
  * UniK will download a 3.3GB vmdk file from Amazon S3. This file is the boot image for the Unik Instance Listener. The instance listener is a small Ubuntu VM that helps bootstrap UniK instances running on Virtualbox.
  * After UniK finishes downloading the vmdk, it will deploy the Instance Listener VM on virtualbox
  * After this is finished, UniK is running and ready to accept commands.

---

#### Write a Go HTTP server
0. Open a new terminal window, but leave the window with the daemon running. This window will be used for running UniK CLI commands.
1. Create a file `httpd.go` using a text editor. Copy and paste the following code in that file:

  ```
  package main

  import (
      "fmt"
      "net/http"
  )

  func main() {
      http.HandleFunc("/", handler)
      http.ListenAndServe(":8080", nil)
  }

  func handler(w http.ResponseWriter, r *http.Request) {
      fmt.Fprintf(w, "my first unikernel!")
  }
  ```
2. Try running this code with `go run http.go`. Visit [http://localhost:8080/](http://localhost:8080/) to see that the server is running.
3. Great! Now we're ready to compile this code to a unikernel.

---

#### Compile an image and run on Virtualbox

1. run the following command from the directory where your `httpd.go` is located:
  ```
  unik build --name myImage --path ./ --compiler rump-go-virtualbox --provider virtualbox
  ```
  this command will instruct UniK to compile the sources found in the working directory (`./`) using the `rump-go-virtualbox` compiler, and stage the image for running the `virtualbox` provider.
2. You can watch the output of the `build` command in the terminal window running the daemon.
3. When `build` finishes, the resulting disk image will reside at `$HOME/.unik/virtualbox/images/myImage/boot.vmdk`
4. Run an instance of this image with
  ```
  unik run --instanceName myInstance --imageName myImage
  ```
5. When the instance finishes launching, let's check its IP and see that it is running our application.
6. Run `unik instances`. The instance IP Address should be listed.
7. Direct your browser to `http://instance-ip:8080` and see that your instance is running!
8. To clean up your image and the instance you created
  ```
  unik rmi --force --image myImage
  ```
