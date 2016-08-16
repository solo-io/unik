# Xen Provider
UniK supports running unikernels on Xen.
In order to run on Xen, you must have `xen-hypervisor-amd64` installed, and `xl` must be a valid command. Running Xen commands with `unik` may require launching the daemon as root.

Currently UniK is configured to run only paravirtualized (not HVM) VMs with Xen.

To run UniK instances with Xen, add a Xen stub to your `daemon-config.yaml`:

```yaml
providers:
  #...
  xen:
    - name: my-xen
      xen_bridge: xenbr0
      pv_kernel: /home/ubuntu/xen/dist/install/usr/local/lib/xen/boot/pv-grub-x86_64.gz
```

`xen_bridge` specifies the name of the bridged interface configured for use with Xen. If you don't have a xen bridge set up, see the instructions at https://help.ubuntu.com/community/Xen.

`pv_kernel` specifies the path to a pv grub boot manager. To install pv-grub, follow the instructions here: https://wiki.xen.org/wiki/PvGrub#Build
