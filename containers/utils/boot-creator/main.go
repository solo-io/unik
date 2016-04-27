package main

import (
	"flag"
	"path"

	log "github.com/Sirupsen/logrus"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
)

func main() {
	log.SetLevel(log.DebugLevel)

	buildcontextdir := flag.String("d", "/opt/vol", "build context. relative volume names are relative to that")
	kernelInContext := flag.String("p", "program.bin", "kernel binary name.")
	args := flag.String("a", "", "arguments to kernel")

	flag.Parse()

	kernelFile := path.Join(*buildcontextdir, *kernelInContext)
	imgFile := path.Join(*buildcontextdir, "vol.img")

	log.WithFields(log.Fields{"kernelFile": kernelFile, "args": *args, "imgFile": imgFile}).Debug("calling CreateBootImageWithSize")

	err := unikos.CreateBootImageWithSize(imgFile, unikos.MegaBytes(100), kernelFile, *args)

	if err != nil {
		log.Fatal(err)
	}
}
