package qemu

import (
	"os"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
)

func (p *QemuProvider) Stage(params types.StageImageParams) (_ *types.Image, err error) {
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
	imagePath := getImagePath(params.Name)
	logrus.Debugf("making directory: %s", filepath.Dir(imagePath))
	if err := os.MkdirAll(filepath.Dir(imagePath), 0777); err != nil {
		return nil, errors.New("creating directory for boot image", err)
	}
	defer func() {
		if err != nil && !params.NoCleanup {
			os.RemoveAll(filepath.Dir(imagePath))
		}
	}()

	logrus.WithField("raw-image", params.RawImage).Infof("creating boot volume from raw image")
	if err := common.ConvertRawImage(params.RawImage.StageSpec.ImageFormat, types.ImageFormat_QCOW2, params.RawImage.LocalImagePath, imagePath); err != nil {
		return nil, errors.New("converting raw image to qcow2", err)
	}

	kernelFile := filepath.Join(filepath.Dir(params.RawImage.LocalImagePath), "program.bin")
	if err := unikos.CopyFile(kernelFile, getKernelPath(params.Name)); err != nil {
		return nil, errors.New("copying kernel file to image dir", err)
	}

	cmdlineFile := filepath.Join(filepath.Dir(params.RawImage.LocalImagePath), "cmdline")
	if err := unikos.CopyFile(cmdlineFile, getCmdlinePath(params.Name)); err != nil {
		return nil, errors.New("copying cmdline file to image dir", err)
	}

	imagePathInfo, err := os.Stat(imagePath)
	if err != nil {
		return nil, errors.New("statting raw image file", err)
	}
	sizeMb := imagePathInfo.Size() >> 20

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
		Infrastructure: types.Infrastructure_QEMU,
		Created:        time.Now(),
	}

	if err := p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[params.Name] = image
		return nil
	}); err != nil {
		return nil, errors.New("modifying image map in state", err)
	}
	if err := p.state.Save(); err != nil {
		return nil, errors.New("saving image map to state", err)
	}

	logrus.WithFields(logrus.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}
