package ukvm

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/solo-io/unik/pkg/os"
	"github.com/solo-io/unik/pkg/types"
)

func (p *UkvmProvider) Stage(params types.StageImageParams) (_ *types.Image, err error) {
	images, err := p.ListImages()
	if err != nil {
		return nil, errors.New("retrieving image list for existing image", err)
	}
	for _, image := range images {
		if image.Name == params.Name {
			if !params.Force {
				return nil, errors.New("an image already exists with name '"+params.Name+"', try again with --force", nil)
			} else {
				logrus.WithField("image", image).Warnf("force: deleting previous image with name " + params.Name)
				if err := p.DeleteImage(image.Id, true); err != nil {
					logrus.Warn("failed to remove previously existing image", err)
				}
			}
		}
	}
	imageName := params.Name
	imageDir := getImageDir(imageName)
	logrus.Debugf("making directory: %s", imageDir)
	if err := os.MkdirAll(imageDir, 0777); err != nil {
		return nil, errors.New("creating directory for boot image", err)
	}
	defer func() {
		if err != nil && !params.NoCleanup {
			os.RemoveAll(imageDir)
		}
	}()

	kernelPath := filepath.Join(params.RawImage.LocalImagePath, "program.bin")
	if err := unikos.CopyFile(kernelPath, getKernelPath(imageName)); err != nil {
		return nil, errors.New("program.bin cannot be copied", err)

	}
	ukvmPath := filepath.Join(params.RawImage.LocalImagePath, "ukvm-bin")
	if err := unikos.CopyFile(ukvmPath, getUkvmPath(imageName)); err != nil {
		return nil, errors.New("ukvm-bin cannot be copied", err)
	}

	kernelPathInfo, err := os.Stat(kernelPath)
	if err != nil {
		return nil, errors.New("statting unikernel file", err)
	}
	ukvmPathInfo, err := os.Stat(ukvmPath)
	if err != nil {
		return nil, errors.New("statting ukvm file", err)
	}
	sizeMb := (ukvmPathInfo.Size() + kernelPathInfo.Size()) >> 20

	logrus.WithFields(logrus.Fields{
		"name": params.Name,
		"id":   params.Name,
		"size": sizeMb,
	}).Infof("copying raw boot image")

	image := &types.Image{
		Id:             params.Name,
		Name:           params.Name,
		RunSpec:        params.RawImage.RunSpec,
		StageSpec:      params.RawImage.StageSpec,
		SizeMb:         sizeMb,
		Infrastructure: types.Infrastructure_UKVM,
		Created:        time.Now(),
	}

	if err := p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[params.Name] = image
		return nil
	}); err != nil {
		return nil, errors.New("modifying image map in state", err)
	}

	logrus.WithFields(logrus.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}
