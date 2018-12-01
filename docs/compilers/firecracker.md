

# For devs - build unik
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

# Create the demo program

The current compiler supports go 1.11. Let's prepare a demo program:
```
mkdir demo
cat > demo/main.go <<EOF
package main

import (
  "fmt"
  "os/exec"
  "time"
)

func main() {
  for {
    fmt.Println("Hello from firecracker (run by unik from solo.io)")
    out, _ := exec.Command("uname", "-a").CombinedOutput()
    fmt.Printf("OS Version: %s\n", string(out))
    time.Sleep(10 * time.Second)
  }
}
EOF
```

# Run unik and build image

In one terminal, run daemon:
```
./unik daemon --debug
```

In other terminal, build and run:
```
./unik build --name writeOS_miVM --path ./demo/ --base firecracker --language go --provider firecracker --force

./unik run --instanceName writeOS_vm1 --imageName writeOS_miVM
```


# Cleanup:
```
./unik delete-instance --instance writeOS_vm1
./unik delete-image --image myImage
```