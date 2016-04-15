package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/config"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"flag"
	"github.com/emc-advanced-dev/unik/pkg/providers/vsphere"
)

func main() {
	action := flag.String("action", "all", "what to test")
	arg := flag.String("arg", "", "option for some test (i.e. instance id)")
	flag.Parse()
	os.Setenv("TMPDIR", os.Getenv("HOME")+"/tmp/uniktest")
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(&uniklog.AddTraceHook{true})

	c := config.Vsphere{
		Name: "vsphere-provider",
		VsphereURL: os.Getenv("VSPHERE_URL"),
		VsphereUser: os.Getenv("VSPHERE_USER"),
		VspherePassword: os.Getenv("VSPHERE_PASSWORD"),
	}
	p, err := vsphere.NewVsphereProvier(c)
	if err != nil {
		logrus.Error(err)
		return
	}
	state, err := state.LocalStorageStateFromFile(vsphere.VsphereStateFile)
	if err != nil {
		logrus.WithError(err).Error("failed to load state")
	} else {
		logrus.Info("state loaded")
		p = p.WithState(state)
	}

	compilers.CreateImageVmware()

	switch(*action){
	case "all":
		r := compilers.RunmpCompiler{
			DockerImage: "rumpcompiler-go-xen",
			CreateImage: compilers.CreateImageAws,
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

		img, err := p.Stage("test-scott", rawimg, true)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("image", img).Infof("printing image")
		fmt.Println()

		env := make(map[string]string)
		env["FOO"] = "BAR"

		instance, err := p.RunInstance("test-scott-instance-1", img.Id, nil, env)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("instance", instance).Infof("printing instance")
		fmt.Println()

		images, err := p.ListImages()
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("images", images).Infof("printing images")
		fmt.Println()

		instances, err := p.ListInstances()
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("instances", instances).Infof("printing instances")
		fmt.Println()

		for _, instance := range instances {
			if err := p.DeleteInstance(instance.Id); err != nil {
				logrus.Error(err)
				return
			}
		}

		for _, image := range images {
			if err := p.DeleteImage(image.Id, false); err != nil {
				logrus.Error(err)
				return
			}
		}
		break
	case "list-images":
		images, err := p.ListImages()
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("images", images).Infof("printing images")
		break
	case "list-instances":
		instances, err := p.ListInstances()
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.WithField("instances", instances).Infof("printing instances")
		break
	case "delete-instance":
		instanceId := *arg
		err = p.DeleteInstance(instanceId)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Infof("deleted instance %s", instanceId)
		break
	case "delete-image":
		imageId := *arg
		err = p.DeleteImage(imageId, true)
		if err != nil {
			logrus.Error(err)
			return
		}
		logrus.Infof("deleted image %s", imageId)
		break
	}

}