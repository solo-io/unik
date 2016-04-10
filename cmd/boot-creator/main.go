// +build linux

package main

import (
	"flag"
	"path"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
)


func main() {
	buildcontextdir := flag.String("d", "/opt/vol", "build context. relative volume names are relative to that")
	kernelInContext := flag.String("p", "program.bin", "kernel binary name.")
	args := flag.String("a", "", "arguments to kernel")

    kernelFile := path.Join(*buildcontextdir, *kernelInContext)
	imgFile := ""

		err := unikos.CreateBootImageOnFile(imgFile, unikos.MegaBytes(100), kernelFile, *args)

	if err != nil {
		panic(err)
	}
}
