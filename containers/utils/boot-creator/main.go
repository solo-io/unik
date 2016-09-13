package main

import (
	"flag"
	"path"

	log "github.com/Sirupsen/logrus"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/pborman/uuid"
	"io"
	"os"
)

const staticFileDir = "/tmp/staticfiles"

func main() {
	log.SetLevel(log.DebugLevel)
	buildcontextdir := flag.String("d", "/opt/vol", "build context. relative volume names are relative to that")
	kernelInContext := flag.String("p", "program.bin", "kernel binary name.")
	usePartitionTables := flag.Bool("part", true, "indicates whether or not to use partition tables and install grub")
	args := flag.String("a", "", "arguments to kernel")
	out := flag.String("o", "", "base name of output file")

	flag.Parse()

	kernelFile := path.Join(*buildcontextdir, *kernelInContext)
	imgFile := path.Join(*buildcontextdir, "boot.image."+uuid.New())
	defer os.Remove(imgFile)

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

	src, err := os.Open(imgFile)
	if err != nil {
		log.Fatal("failed to open produced image file "+imgFile, err)
	}
	outFile := path.Join(*buildcontextdir, *out)
	dst, err := os.OpenFile(outFile, os.O_RDWR, 0)
	if err != nil {
		log.Fatal("failed to open target output file "+outFile, err)
	}
	n, err := io.Copy(dst, src)
	if err != nil {
		log.Fatal("failed copying produced image file to target output file", err)
	}
	log.Info("wrote %d bytes to disk", n)
}
