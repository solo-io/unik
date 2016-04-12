package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/daemon"
	"os"
)

func main() {
	logrus.Info("before the thing")
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Debugf("before the thing debug style")
	logrus.AddHook(daemon.NewUnikLogrusHook(os.Stdout, "test-logger"))
	logrus.Infof("i keep my nails clean %s", "not really")
	logrus.WithFields(logrus.Fields{"here's some info": 1}).Warnf("i keep my nails clean %s", "not really")
}