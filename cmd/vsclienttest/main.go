package main

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/vsphere/vsphereclient"
	"os"
	"net/url"
	"github.com/Sirupsen/logrus"
	"time"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"path/filepath"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"flag"
)

func main() {
	destroy := flag.Bool("destroy", false, "destroy evverything")
	copyVmdk := flag.Bool("copy", false, "copy instance listener vmdk")
	flag.Parse()
	os.Setenv("TMPDIR", os.Getenv("HOME")+"/tmp/uniktest")
	rawUrl := os.Getenv("VSPHERE_URL")
	u, err := url.Parse(rawUrl)
	if err != nil {
		logrus.Panic(err)
	}

	deferred := util.Stack{}
	defer func(){
		for fn := deferred.Pop(); fn != nil; {
			fn.(func())()
		}
	}()

	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(&uniklog.AddTraceHook{true})
	c := vsphereclient.NewVsphereClient(u, "datastore1")

	if *destroy {
		logrus.Infof("TEARING DOWN!")
		logrus.Infof("rmdir dir uniktest")
		if err := c.Rmdir("uniktest"); err != nil {
			logrus.Error(err)
		}
		logrus.Infof("detaching boot.vmdk from uniktest-vm")
		if err := c.DetachDisk("uniktest-vm", 0); err != nil {
			logrus.Error(err)
		}
		logrus.Infof("powering off uniktest-vm")
		if err := c.PowerOffVm("uniktest-vm"); err != nil {
			logrus.Error(err)
		}
		logrus.Infof("destroy vm uniktest-vm")
		if err := c.DestroyVm("uniktest-vm"); err != nil {
			logrus.Error(err)
		}

		return
	}
	if *copyVmdk {
		logrus.Infof("copying vmdk")
		if err := c.CopyVmdk("unik/instancelistener-base.vmdk", "fakedir/instancelistener-copy.vmdk"); err != nil {
			logrus.Error(err)
		}
		return
	}


	logrus.Infof("making dir uniktest")
	if err := c.Mkdir("uniktest"); err != nil {
		logrus.Panic(err)
	}
	deferred.Push(func(){
		logrus.Infof("rmdir dir uniktest")
		if err := c.Rmdir("uniktest"); err != nil {
			logrus.Panic(err)
		}
	})

	logrus.Infof("create vm uniktest-vm")
	if err := c.CreateVm("uniktest-vm", 512); err != nil {
		logrus.Panic(err)
	}
	deferred.Push(func(){
		logrus.Infof("destroy vm uniktest-vm")
		if err := c.DestroyVm("uniktest-vm"); err != nil {
			logrus.Panic(err)
		}
	})

	logrus.Infof("importing boot.vmdk to uniktest-vm folder")
	vmdkPath, err := filepath.Abs("./boot.vmdk")
	if err != nil {
		logrus.Panic(err)
	}
	if err := c.ImportVmdk(vmdkPath, "uniktest-vm"); err != nil {
		logrus.Panic(err)
	}

	logrus.Infof("attaching boot.vmdk to uniktest-vm")
	if err := c.AttachDisk("uniktest-vm", "uniktest-vm/boot.vmdk", 0); err != nil {
		logrus.Panic(err)
	}
	deferred.Push(func(){
		logrus.Infof("detaching boot.vmdk from uniktest-vm")
		if err := c.DetachDisk("uniktest-vm", 0); err != nil {
			logrus.Panic(err)
		}
	})

	logrus.Infof("powering on uniktest-vm")
	if err := c.PowerOnVm("uniktest-vm"); err != nil {
		logrus.Panic(err)
	}
	deferred.Push(func(){
		logrus.Infof("powering off uniktest-vm")
		if err := c.PowerOffVm("uniktest-vm"); err != nil {
			logrus.Panic(err)
		}
	})

	logrus.Infof("get uniktest-vm")
	if vm, err := c.GetVm("uniktest-vm"); err != nil {
		logrus.Panic(err)
	} else {
		logrus.Infof("GOT vm:\n%v", vm)
	}
	for i := 15; i > 0; i-- {
		logrus.Infof("vm is running! go check it out! %v seconds left...", i)
		time.Sleep(time.Second)
	}
	logrus.Infof("getting vm ip")
	if ip, err := c.GetVmIp("UnikAppliance"); err != nil {
		logrus.Panic(err)
	} else {
		logrus.Infof("ip: %s", ip)
	}
}
