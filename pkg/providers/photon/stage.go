package photon

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

func createVmdk(params types.StageImageParams, workVmdk func(file string) (string, error)) (string, int64, error) {

	localVmdkDir, err := ioutil.TempDir("", "vmdkdir.")
	if err != nil {
		return "", 0, errors.New("creating tmp file", err)
	}
	defer os.RemoveAll(localVmdkDir)
	localVmdkFile := filepath.Join(localVmdkDir, "boot.vmdk")

	logrus.WithField("raw-image", params.RawImage).Infof("creating boot volume from raw image")
	if err := common.ConvertRawToNewVmdk(params.RawImage.LocalImagePath, localVmdkFile); err != nil {
		return "", 0, errors.New("converting raw image to vmdk", err)
	}

	rawImageFile, err := os.Stat(localVmdkFile)
	if err != nil {
		return "", 0, errors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	logrus.WithFields(logrus.Fields{
		"name": params.Name,
		"id":   params.Name,
		"size": sizeMb,
	}).Infof("importing base vmdk for unikernel image")

	imgId, err := workVmdk(localVmdkFile)
	return imgId, sizeMb, err

}

func (p *PhotonProvider) Stage(params types.StageImageParams) (_ *types.Image, err error) {
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
					logrus.Warn(errors.New("failed removing previously existing image", err))
				}
			}
		}
	}

	// create vmdk
	imgId, sizeMb, err := createVmdk(params, func(vmdkFile string) (string, error) {
		options := &photon.ImageCreateOptions{
			ReplicationType: "EAGER",
		}
		task, err := p.client.Images.CreateFromFile(vmdkFile, options)
		if err != nil {
			return "", errors.New("error creating photon image", err)
		}

		task, err = p.waitForTaskSuccess(task)
		if err != nil {
			return "", errors.New("error waiting for task creating photon image", err)
		}

		return task.Entity.ID, nil
	})
	if err != nil {
		return nil, errors.New("importing base boot.vmdk to photon", err)
	}

	// upload images
	image := &types.Image{
		Id:               imgId,
		Name:             params.Name,
		StageSpec:        params.RawImage.StageSpec,
		RunSpec:          params.RawImage.RunSpec,
		SizeMb:           sizeMb,
		Infrastructure:   types.Infrastructure_PHOTON,
		Created:          time.Now(),
	}

	err = p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[params.Name] = image
		return nil
	})
	if err != nil {
		return nil, errors.New("modifying image map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return nil, errors.New("saving image map to state", err)
	}

	logrus.WithFields(logrus.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}
