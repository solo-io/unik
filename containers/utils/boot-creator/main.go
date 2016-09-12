package main

import (
	"flag"
	"path"

	log "github.com/Sirupsen/logrus"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"os"
)

const staticFileDir = "/tmp/staticfiles"

func main() {
	log.SetLevel(log.DebugLevel)
	buildcontextdir := flag.String("d", "/opt/vol", "build context. relative volume names are relative to that")
	kernelInContext := flag.String("p", "program.bin", "kernel binary name.")
	usePartitionTables := flag.Bool("part", true, "indicates whether or not to use partition tables and install grub")
	strictMode := flag.Bool("strict", false, "disable automatic chmod a+rw on output file (fixes issue #40)")
	args := flag.String("a", "", "arguments to kernel")

	flag.Parse()

	kernelFile := path.Join(*buildcontextdir, *kernelInContext)
	imgFile := path.Join(*buildcontextdir, "vol.img")

	log.WithFields(log.Fields{"kernelFile": kernelFile, "args": *args, "imgFile": imgFile, "usePartitionTables": *usePartitionTables}).Debug("calling CreateBootImageWithSize")

	s1, err := unikos.DirSize(*buildcontextdir)
	if err != nil {
		log.Fatal(err)
	}
	s2 := float64(s1) * 1.1
	size := ((int64(s2) >> 20) + 10)

	if err := unikos.CopyDir(*buildcontextdir, staticFileDir); err != nil {
		log.Fatal(err)
	}

	//no need to copy twice
	os.Remove(path.Join(staticFileDir, *kernelInContext))

	if err := unikos.CreateBootImageWithSize(imgFile, unikos.MegaBytes(size), kernelFile, staticFileDir, *args, *usePartitionTables); err != nil {
		log.Fatal(err)
	}

	if !*strictMode {
		info, err := os.Stat(imgFile)
		if err != nil {
			log.Fatal("could not stat image file "+imgFile, err)
		}
		if err := os.Chmod(imgFile, info.Mode()|0666); err != nil {
			log.Fatal("adding rw permission to image file", err)
		}
	}
}
