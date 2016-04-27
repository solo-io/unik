package main

import (
	"flag"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/config"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"os"
)

func main() {
	operation := flag.String("op", "list", "action")
	flag.Parse()

	baseFolder := os.Getenv("PWD")
	//hostnetworkName := "en0"
	hostnetworkName := "vboxnet0"
	diskFile := "./boot.vmdk"
	logrus.SetLevel(logrus.DebugLevel)
	switch *operation {
	case "list":
		vms, err := virtualboxclient.Vms()
		if err != nil {
			logrus.WithError(err).Fatalf("getting vm list")
		}
		logrus.WithField("vms", vms).Info("get vms succeeded")
	case "create-vm":
		err := virtualboxclient.CreateVm("test-scott", baseFolder, hostnetworkName, config.HostOnlyAdapter)
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
	case "detach-disk":
		err := virtualboxclient.DetachDisk("test-scott", 0)
		if err != nil {
			logrus.WithError(err).Fatalf("detaching disk to vm")
		}
	case "detach-data-disk":
		err := virtualboxclient.DetachDisk("test-scott", 1)
		if err != nil {
			logrus.WithError(err).Fatalf("detaching disk to vm")
		}
	case "get-vm-ip":
		ip, err := virtualboxclient.GetVmIp(virtualbox.VboxUnikInstanceListener)
		if err != nil {
			logrus.WithError(err).Fatalf("getting vm ip")
		}
		logrus.WithField("ip", ip).Info("get ip succeeded")
	case "create-instance-listener":
		if err := virtualboxclient.CreateVmNatless(virtualbox.VboxUnikInstanceListener, baseFolder, hostnetworkName, config.HostOnlyAdapter); err != nil {
			logrus.WithError(err).Fatalf("creating vm")
		}
		if err := unikos.CopyFile("instancelistener-base.vmdk", "instancelistener-copy.vmdk"); err != nil {
			logrus.WithError(err).Fatalf("copying instance listener vmdk")
		}
		if err := virtualboxclient.AttachDisk(virtualbox.VboxUnikInstanceListener, "instancelistener-copy.vmdk", 0); err != nil {
			logrus.WithError(err).Fatalf("attaching disk to vm")
		}
		if err := virtualboxclient.PowerOnVm(virtualbox.VboxUnikInstanceListener); err != nil {
			logrus.WithError(err).Fatalf("powering on vm")
		}
	case "destroy-instance-listener":
		err := virtualboxclient.PowerOffVm(virtualbox.VboxUnikInstanceListener)
		if err != nil {
			logrus.WithError(err).Fatalf("powering off vm")
		}
		err = virtualboxclient.DestroyVm(virtualbox.VboxUnikInstanceListener)
		if err != nil {
			logrus.WithError(err).Fatalf("destroying vm")
		}
	}

}
