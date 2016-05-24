package main

import (
	"flag"
	"path"

	log "github.com/Sirupsen/logrus"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	//unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"os"
)

func main() {
	log.SetLevel(log.DebugLevel)
	//log.AddHook(&unikutil.AddTraceHook{true})
	buildcontextdir := flag.String("d", "/opt/vol", "build context. relative volume names are relative to that")
	kernelInContext := flag.String("p", "program.bin", "kernel binary name.")
	args := flag.String("a", "", "arguments to kernel")

	flag.Parse()

	kernelFile := path.Join(*buildcontextdir, *kernelInContext)
	imgFile := path.Join(*buildcontextdir, "vol.img")

	log.WithFields(log.Fields{"kernelFile": kernelFile, "args": *args, "imgFile": imgFile}).Debug("calling CreateBootImageWithSize")

	kernelFileInfo, err := os.Stat(kernelFile)
	if err != nil {
		log.Fatal(err)
	}
	s1 := float64(kernelFileInfo.Size()) * 1.1
	size := (int64(s1) >> 20) + 10

	if err := unikos.CreateBootImageWithSize(imgFile, unikos.MegaBytes(size), kernelFile, *args); err != nil {
		log.Fatal(err)
	}
}
