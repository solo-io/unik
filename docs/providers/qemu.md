# QEMU Provider
UniK supports running rumprun unikernels through QEMU.
In order to run on QEMU, you must have `qemu` installed.

To run UniK instances with QEMU, add a QEMU stub to your `daemon-config.yaml`:

```yaml
providers:
  #...
  qemu:
    - name: my-qemu
      no_graphic: false
```

`no_graphic` specifies whether or not QEMU instances will be launched using a `no-graphic` mode. Set to `true` for environments with no desktop/graphical interface.

As QEMU is not a full hypervisor, the QEMU provider has some limitations, and is ideal mostly for debugging unikernels.

The QEMU provider supports the `--debug-mode` option for running unikernels, which will launch a unikernel in *stopped* mode and attach [`gdb`](https://www.gnu.org/software/gdb/) remotely to the unikernel, allowing line-by-line debugging of the source code for the unikernel.

Limitations of QEMU provider:
* Instances cannot be powered down. Powering down an instance will terminate it. Killing the UniK Daemon will terminate all QEMU instances, but they will still have to be deleted from UniK's state with `unik rm --instance <instance_name>` in order for UniK to know they are no longer running.
* QEMU instances will be assigned IPs and will have network connectivity, but will not be reachable from the host network. It is possible to configure a `tap` device with a bridge to enable instances to be reachable, but we are not supporting this feature at this time.
* QEMU instances do not make use of the UniK bootstrapping stub/wrapper.
