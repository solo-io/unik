package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"os"
)

func main() {
	op := flag.String("op", "boot", "creates boot|data image")
	flag.Parse()
	logrus.SetLevel(logrus.DebugLevel)
	os.Setenv("TMPDIR", os.Getenv("HOME")+"/tmp/uniktest")
	f, err := os.Open("a.tar")
	if err != nil {
		logrus.Error(err)
		return
	}
	defer f.Close()
	switch *op {
	case "boot":
		r := compilers.RumpCompiler{
			DockerImage: "compilers-rump-go-hw",
			CreateImage: compilers.CreateImageVirtualBox,
		}
		rawimg, err := r.CompileRawImage(f, "", []string{"/data1", "/data2"})
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("image", rawimg).Infof("image completed")
		break
	case "data":
		imagePath, err := unikos.BuildRawDataImage(f, 0, true)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("image", imagePath).Infof("image completed")
	}
}
