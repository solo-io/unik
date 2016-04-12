package main

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/providers/aws"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"time"
	uniklog "github.com/emc-advanced-dev/unik/pkg/util/log"
)

func main() {
	os.Setenv("TMPDIR", "/Users/pivotal/tmp/uniktest")
	logrus.SetLevel(logrus.DebugLevel)
	logrus.AddHook(&uniklog.AddTraceHook{true})

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
	c := config.Aws{
		Name: "aws-provider",
		AwsAccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"),
		AwsSecretAcessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region: os.Getenv("AWS_REGION"),
		Zone: os.Getenv("AWS_AVAILABILITY_ZONE"),
	}
	p := aws.NewAwsProvier(c)
	err = p.Load()
	if err != nil {
		logrus.WithError(err).Error("failed to load state")
	} else {
		logrus.Info("state loaded")
	}
	saveState(p)
	defer func() {
		saveState(p)
	}()

	go func(){
		for {
			saveState(p)
			time.Sleep(5000 * time.Millisecond)
		}
	}()

	img, err := p.Stage("test-scott", rawimg, true)
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Print(img)
	fmt.Println()

	env := make(map[string]string)
	env["FOO"] = "BAR"

	instance, err := p.RunInstance("test-scott-instance-1", img.Id, nil, env)
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Print(instance)
	fmt.Println()

	images, err := p.ListImages()
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Print(images)
	fmt.Println()

	instances, err := p.ListInstances()
	if err != nil {
		logrus.Error(err)
		return
	}
	fmt.Print(instances)
	fmt.Println()

	for _, instance := range instances {
		err = p.DeleteInstance(instance.Id)
		if err != nil {
			logrus.Error(err)
			return
		}
	}

	for _, image := range images {
		err = p.DeleteImage(image.Id, false)
		if err != nil {
			logrus.Error(err)
			return
		}
	}
}

func saveState(p *aws.AwsProvider) {
	err := p.Save()
	if err != nil {
		logrus.WithError(err).Fatalf("failed to save")
	} else {
		logrus.Info("saved state")
	}
}