package main

import (
	//"os"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/Sirupsen/logrus"
)

func main() {
	/*diskPath := "/Users/pivotal/VirtualBox VMs/Windows10/Windows10.vbox"
	baseFolder := os.Getenv("PWD")
	bridgeName := "bridgestuff"
	bridgeAdapterKey := 0
	diskFile := "./boot.vmdk"*/
	logrus.SetLevel(logrus.DebugLevel)
	vms, err := virtualboxclient.Vms()
	if err != nil {
		logrus.WithError(err).Panic("getting vm list")
	}
	logrus.WithField("vms", vms).Info("get vms succeeded")
}