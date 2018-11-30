package main

import (
	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"os"
)

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {
	logrus.SetLevel(logrus.DebugLevel)
	args := os.Args[1:]
	if len(args) != 4 {
		logrus.Warnf("bad args: %v", args)
		return errors.New("usage: go run main.go PROJECT IMAGE BUCKET OBJECT", nil)
	}
	project := args[0]
	name := args[1]
	bucket := args[2]
	object := args[3]

	// Use oauth2.NoContext if there isn't a good context to pass in.
	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope, compute.CloudPlatformScope, compute.DevstorageReadWriteScope)
	if err != nil {
		return errors.New("failed to start default client", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return errors.New("failed to start compute client", err)
	}

	imageSpec := &compute.Image{
		Name: name,
		RawDisk: &compute.ImageRawDisk{
			ContainerType: "TAR",
			Source:        "http://storage.googleapis.com/" + bucket + "/" + object,
		},
		SourceType: "RAW",
	}

	logrus.Debugf("creating image from " + imageSpec.RawDisk.Source)

	req := computeService.Images.Insert(project, imageSpec)

	logrus.Debugf("sending request:\n%+v", req)

	gImage, err := req.Do()
	if err != nil {
		return errors.New("creating gcloud image from storage", err)
	}
	logrus.Infof("success: %v", gImage)
	return nil
}
