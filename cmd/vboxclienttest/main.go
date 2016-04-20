package main

import (
	//"os"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/Sirupsen/logrus"
	"os"
	"flag"
)

func main() {
	operation := flag.String("op", "list", "action")
	flag.Parse()

	baseFolder := os.Getenv("PWD")
	bridgeName := "en0"
	diskFile := "./boot.vmdk"
	logrus.SetLevel(logrus.DebugLevel)
	switch(*operation){
	case "list":
		vms, err := virtualboxclient.Vms()
		if err != nil {
			logrus.WithError(err).Fatalf("getting vm list")
		}
		logrus.WithField("vms", vms).Info("get vms succeeded")
	case "create-vm":
		err := virtualboxclient.CreateVm("test-scott", baseFolder, bridgeName)
		if err != nil {
			logrus.WithError(err).Fatalf("creating vm")
		}
	case "destroy-vm":
		err := virtualboxclient.DestroyVm("test-scott")
		if err != nil {
			logrus.WithError(err).Fatalf("destroying vm")
		}
	case "power-on":
		err := virtualboxclient.PowerOnVm("test-scott")
		if err != nil {
			logrus.WithError(err).Fatalf("powering on vm")
		}
	case "power-off":
		err := virtualboxclient.PowerOffVm("test-scott")
		if err != nil {
			logrus.WithError(err).Fatalf("powering off vm")
		}
	case "attach-disk":
		err := virtualboxclient.AttachDisk("test-scott", diskFile, 0)
		if err != nil {
			logrus.WithError(err).Fatalf("attaching disk to vm")
		}
	case "attach-data-disk":
		err := virtualboxclient.AttachDisk("test-scott", "./data.vmdk", 1)
		if err != nil {
			logrus.WithError(err).Fatalf("attaching disk to vm")
		}
	}

}