

# For devs - build unikq
If needed (during dev cycle), build compiler and unik:
```
make compilers-firecracker
make localbuild
```

# Config

Follow regular getting started instructions, and configure the firecracker provider like so:

```
providers:
  firecracker:
    - name: firecracker
      binary: /path/to/firecracker
      kernel: /path/to/kernel/hello-vmlinux.bin
      console: xterm
```

# Run unik and build image

In one terminal, run daemon:
```
./unik daemon --debug
```

In other terminal, build and run:
```
./unik build --name myImage --path ./t/ --base firecracker --language go --provider firecracker --force
./unik run --instanceName myInstance --imageName myImage
```


# Cleanup:
```
./unik delete-instance --instance myInstance
./unik delete-image --image myImage
```