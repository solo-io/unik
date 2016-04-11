package main

import (
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/providers/aws"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/layer-x/layerx-commons/lxlog"
)

func main() {
	os.Setenv("TMPDIR", "/Users/pivotal/tmp/uniktest")
	log.SetLevel(log.DebugLevel)

	r := compilers.RunmpCompiler{
		DockerImage: "rumpcompiler-go-xen",
		CreateImage: compilers.CreateImageAws,
	}
	f, err := os.Open("a.tar")
	if err != nil {
		panic(err)
	}
	rawimg, err := r.CompileRawImage(f, "", []string{"/yuval"})
	if err != nil {
		panic(err)
	}
	c := config.Aws{
		Name: "aws-provider",
		AwsAccessKeyID: os.Getenv("AWS_ACCESS_KEY_ID"),
		AwsSecretAcessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region: os.Getenv("AWS_REGION"),
		Zone: os.Getenv("AWS_AVAILABILITY_ZONE"),
	}
	p := aws.NewAwsProvier(c)

	logger := lxlog.New("scott")

	img, err := p.Stage(logger, "test-scott", rawimg, true)
	if err != nil {
		panic(err)
	}


	fmt.Print(img)
}
