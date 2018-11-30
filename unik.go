package main

import (
	"github.com/sirupsen/logrus"
	"github.com/solo-io/unik/cmd"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		logrus.WithError(err).Fatalf("unik failed")
	}
}
