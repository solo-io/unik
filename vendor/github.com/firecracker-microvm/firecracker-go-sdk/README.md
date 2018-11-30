A basic Go interface to the Firecracker API
====

This package is a Go library to interact with the Firecracker API.  There is
also a command-line application that can be used to control Firecracker microVMs
called `firectl`.

There are some Firecracker features that are not yet supported by the SDK.
These are tracked as GitHub issues with the
[firecracker-feature](https://github.com/firecracker-microvm/firecracker-go-sdk/issues?q=is%3Aissue+is%3Aopen+label%3Afirecracker-feature)
label . Contributions to address missing features are welcomed.

Developing
---

Please see [HACKING](HACKING.md)

Building
---

This library requires Go 1.11 and Go modules to build.  A Makefile is provided
for convenience, but is not required.  When using the Makefile, you can pass
additional flags to the Go compiler via the `EXTRAGOARGS` make variable.

Tools
---

There's a basic command-line tool, built as `cmd/firectl/firectl` that lets you
run arbitrary Firecracker MicroVMs via the command line. This lets you run a
fully functional Firecracker MicroVM, including console access, read/write
access to filesystems, and network connectivity.

```
Usage of ./cmd/firectl/firectl:
    --firecracker-binary=  Path to Firecracker binary
    --firecracker-console= Console type (stdio|xterm|none) (default: stdio)
    --kernel=              Path to the kernel image (default: ./vmlinux)
    --kernel-opts=         Kernel commandline (default: ro console=ttyS0 noapic reboot=k panic=1 pci=off nomodules)
    --root-drive=          Path to root disk image
    --add-drive=           Path to additional drive, suffixed with :ro or :rw, can be specified multiple times
    --tap-device=          NIC info, specified as DEVICE:MAC
    --vmm-log-fifo=        FIFO for Firecracker logs
    --log-level=           vmm log level (default: Debug)
    --metrics-fifo=        FIFO for Firecracker metrics
-d, --debug                Enable debug output
-h, --help                 Show usage
```

`$ ./cmd/firectl/firectl --firecracker-binary=./firecracker-0.10.1 --firecracker-console=stdio --root-drive=openwrt-x86-64-rootfs-squashfs.img --tap-device=vmtap33/9a:e4:f6:b0:2d:f3 --add-drive drive-2.img:ro -d --vmm-log-fifo=/tmp/fc-logs.fifo --metrics-fifo=/tmp/fc-metrics`

Network configuration
---

Firecracker, by design, only supports Linux tap devices. The SDK
provides facilities to attach a tap device to the Firecracker VM, but
the client is responsible for further configuration.

License
====

This library is licensed under the Apache 2.0 License. 
