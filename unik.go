package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/cf-unik/unik/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatalf("unik failed")
	}
}
