package main

import (
	"os"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	os.Setenv("TMPDIR", os.Getenv("HOME")+"/tmp/uniktest")
	r := compilers.RunmpCompiler{
		DockerImage: "compilers-rump-go-hw",
		CreateImage: compilers.CreateImageVirtualBox,
	}
	f, err := os.Open("a.tar")
	if err != nil {
		logrus.Error(err)
		return
	}
	rawimg, err := r.CompileRawImage(f, "", []string{})
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.WithField("image", rawimg).Infof("image completed")
}