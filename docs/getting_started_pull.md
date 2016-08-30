# Getting Started

In this tutorial we'll be:
  1. [installing UniK](getting_started_pull.md#installing-unik)
  2. [pull a pre-compiled unikernel image from hub.project-unik.io](getting_started_pull.md#pull-an-existing-unik-unikernel-from-httphubproject-unikio)
  3. [launching an instance on Virtualbox](getting_started_pull.md#run-an-instance-of-the-image-on-virtualbox)

### Installing UniK
#### Prerequisites
Ensure that each of the following are installed
- [Docker](http://www.docker.com/) installed and running with at least 6GB available space for building images
- [`jq`](https://stedolan.github.io/jq/)
- [`make`](https://www.gnu.org/software/make/)
- [Virtualbox](https://www.virtualbox.org/)

#### Install, configure, and launch UniK
1. Install UniK
  ```
  $ git clone https://github.com/emc-advanced-dev/unik.git
  $ cd unik
  $ make
  ```
  note: `make` will take quite a few minutes the first time it runs. the UniK `Makefile` is pulling all of the Docker images that bundle UniK's dependencies.

  Then, place the `unik` executable in your `$PATH` to make running UniK commands easier:
  ```
  $ mv _build/unik /usr/local/bin/
  ```

2. Configure a Host-Only Network on Virtualbox
  * Open Virtualbox
  * Open **Preferences** > **Network** > **Host-only Networks**
  * Click the green add button on the right side of the UI
  * Record the name of the new Host-Only adapter. You will need this in your UniK configuration
  * Ensure that the Virtualbox DHCP Server is Enabled for this Host-Only Network:
    * With the Host-Only Network selected, Click the edit button (screwdriver image)
    * In the **Adapter** tab, note the IPv4 address and netmask of the adapter.
    * In the **DHCP Server** tab, check the **Enable Server** box
    * Set **Server Address** an IP on the same subnet as the Adapter IP. For example, if the adapter IP is `192.168.100.1`, make set the DHCP server IP as `192.168.100.X`, where X is a number between 2-254.
    * Set **Server Mask** to the netmask you just noted
    * Set **Upper / Lower Address Bound** to a range of IPs on the same subnet. We recommend using the range `X-254` where X is one higher than the IP you used for the DHCP server itself. E.g., if your DHCP server is `192.168.100.2`, you can set the lower and upper bounds to `192.168.100.3` and `192.168.100.254`, respectively.


3. Configure UniK daemon
  * Using a text editor, create and save the following to `$HOME/.unik/daemon-config.yaml`:
  ```yaml
  providers:
    virtualbox:
      - name: my-vbox
        adapter_type: host_only
        adapter_name: NEW_HOST_ONLY_ADAPTER
  ```
  replacing `NEW_HOST_ONLY_ADAPTER` with the name of the network adapter you created.


4. Launch UniK and automatically deploy the *Virtualbox Instance Listener*
  * Open a new terminal window/tab. This terminal will be where we leave the UniK daemon running.
  * `cd` to the `_build` directory created by `make`
  * run `unik daemon --debug` (the `--debug` flag is optional, if you want to see more verbose output)
  * UniK will compile and deploy its own 30 MB unikernel. This unikernel is the [Unik Instance Listener](./instance_listener.md). The instance listener uses udp broadcast to detect instance ips and bootstrap instances running on Virtualbox.
  * After this is finished, UniK is running and ready to accept commands.
  * Open a new terminal window and type `unik target --host localhost` to set the CLI target to the your local machine.

---

#### Pull an existing Unik Unikernel from http://hub.project-unik.io

0. Open a new terminal window, but leave the window with the daemon running. This window will be used for running UniK CLI commands.

1. Open another terminal window and type `unik login`
* You will be asked to fill in a URL, Username, and Password. For URL, just hit Enter (this will tell Unik to use the default http://hub.project-unik.io). Pick any username/password you like.
2. List available images with `unik search`
3. Choose an image from the list (by name) and download it to your local storage with `unik pull --image PythonExample --provider virtualbox`
4. Great! Now we're ready to run our first unikernel.

---

#### Run an instance of the image on Virtualbox

4. Run an instance of this image with
  ```
  unik run --instanceName myInstance --imageName PythonExample
  ```
5. When the instance finishes launching, let's check its IP and see that it is running our application.
6. Run `unik instances`. The instance IP Address should be listed.
7. Direct your browser to `http://instance-ip:8080` and see that your instance is running!
8. To clean up your image and the instance you created
  ```
  unik rmi --force --image PythonExample
  ```
