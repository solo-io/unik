package main

import (
	"github.com/Sirupsen/logrus"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"os"
)

func main() {
	os.Setenv("TMPDIR", os.Getenv("HOME")+"/tmp/uniktest")
	logrus.SetLevel(logrus.DebugLevel)
	f, err := os.Open("a.tar")
	if err != nil {
		logrus.Error(err)
		return
	}
	defer f.Close()
	resultFile, err := unikos.BuildRawDataImage(f, 0)
	if err != nil {
		logrus.Error(err)
		return
	}
	logrus.Infof("succeeded: %s", resultFile)
}
