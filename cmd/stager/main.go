package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
)

func main() {

	log.SetLevel(log.DebugLevel)

	r := compilers.RunmpCompiler{
		DockerImage: "rumpcompiler-go-xen",
		CreateImage: compilers.CreateImageAws,
	}
	f, err := os.Open("a.tar")
	if err != nil {
		panic(err)
	}
	img, err := r.CompileRawImage(f, "", []string{})
	if err != nil {
		panic(err)
	}

	fmt.Print(img)
}
