# Debugging status

## VMware
VMware offers a gdb stub that you can use for live gdb debugging, but it is hard to debug on boot, as there is no way to wait for the debugger.

## VirtuaBox
The have their own debugger. I did not find a working gdb stub for 64 bit guests. I only use it when i have to.

## QEMU
Best debugging support, gdb stub and suspends boot.

# Debugging tips

In general whenever I had a problem, I tried to emulate the conditions of the problem in Qemu, as it offers the best debugging experience:
- You can use QEMU to source level debugging.
- Can use different qemu configurations to simulate different hypervisors (this is usefull to test that various drivers work as expected).
- Can wait for debugger to debug boot problems
- Easy command line workflow

# To run QEMU similar to VirtualBox:

Use a scsi driver for disks and two network cards (unik uses two network cards in virtualbox):

    qemu-system-x86_64 -device virtio-scsi-pci,id=scsi \
                   -device scsi-hd,drive=hd1 \
                   -drive file=/Users/kohavy/.unik/virtualbox/images/testboot/boot.vmdk,format=vmdk,if=none,id=hd1 \
                   -device virtio-net-pci,netdev=mynet0,mac=54:54:00:55:55:55 \
                   -netdev user,id=mynet0,net=192.168.76.0/24,dhcpstart=192.168.76.9 \
                   -device virtio-net-pci,netdev=mynet1,mac=54:54:00:55:55:51 \
                   -netdev user,id=mynet1,net=192.168.76.0/24,dhcpstart=192.168.76.9

To see the output of qemu in the console screen, add "-nographic -vga none"

# To run QEMU similar to VMware:

On vmware the network card is behind PCI bridge:

    qemu-system-x86_64 -drive file=root.img,format=raw,if=virtio \
    -device pci-bridge,chassis_nr=2 \
    -device e1000,netdev=mynet0,mac=54:54:00:55:55:55,bus=pci.1,addr=1 \
    -netdev user,id=mynet0,net=192.168.76.0/24,dhcpstart=192.168.76.9

```
qemu-system-x86_64 -device virtio-scsi-pci,id=scsi \
                   -device scsi-hd,drive=hd1 \
                   -drive file=/Users/pivotal/.unik/virtualbox/instances/VboxUnikInstanceListener/boot.vmdk,format=vmdk,if=none,id=hd1 \
                   -device scsi-hd,drive=hd2 \
                   -drive file=/Users/pivotal/.unik/virtualbox/volumes/InstanceListenerData/data.vmdk,format=vmdk,if=none,id=hd2 \
                   -device virtio-net-pci,netdev=mynet0,mac=54:54:00:55:55:55 \
                   -netdev user,id=mynet0,net=192.168.76.0/24,dhcpstart=192.168.76.9 \
                   -device virtio-net-pci,netdev=mynet1,mac=54:54:00:55:55:51 \
                   -netdev user,id=mynet1,net=192.168.76.0/24,dhcpstart=192.168.76.9 \
                   -m 512
```

qemu-system-x86_64 -device virtio-scsi-pci,id=scsi \
                   -device scsi-hd,drive=hd1 \
                   -drive file=/Users/pivotal/.unik/virtualbox/instances/warvbox/boot.vmdk,format=vmdk,if=none,id=hd1 \
                   -device virtio-net-pci,netdev=mynet0,mac=54:54:00:55:55:55 \
                   -netdev user,id=mynet0,net=192.168.76.0/24,dhcpstart=192.168.76.9 \
                   -device virtio-net-pci,netdev=mynet1,mac=54:54:00:55:55:51 \
                   -netdev user,id=mynet1,net=192.168.76.0/24,dhcpstart=192.168.76.9 \
                   -m 1024


For hard drivers, use the scsi drive like in the virtualbox example.

# To debug using gdb

    add "-s -S" to qemu cmdline to enabled debugging.

Use our debugging container:

    docker run --rm -ti --net="host" -v $PWD/:/opt/code projectunik/debuggers-rump-base-hw

and then from inside the container:

    /opt/gdb-7.11/gdb/gdb -ex 'target remote 192.168.99.1:1234' /opt/code/program.bin
