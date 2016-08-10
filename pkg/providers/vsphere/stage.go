package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (p *VsphereProvider) Stage(params types.StageImageParams) (_ *types.Image, err error) {
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
	c := p.getClient()
	vsphereImageDir := getImageDatastoreDir(params.Name)
	if err := c.Mkdir(vsphereImageDir); err != nil && !strings.Contains(err.Error(), "exists") {
		return nil, errors.New("creating vsphere directory for image", err)
	}
	defer func() {
		if err != nil {
			logrus.WithError(err).Warnf("creating image failed, cleaning up image on datastore")
			c.Rmdir(vsphereImageDir)
		}
	}()

	localVmdkDir, err := ioutil.TempDir("", "vmdkdir.")
	if err != nil {
		return nil, errors.New("creating tmp file", err)
	}
	defer os.RemoveAll(localVmdkDir)
	localVmdkFile := filepath.Join(localVmdkDir, "boot.vmdk")

	logrus.WithField("raw-image", params.RawImage).Infof("creating boot volume from raw image")
	if err := common.ConvertRawImage(params.RawImage.StageSpec.ImageFormat, types.ImageFormat_VMDK, params.RawImage.LocalImagePath, localVmdkFile); err != nil {
		return nil, errors.New("converting raw image to vmdk", err)
	}

	rawImageFile, err := os.Stat(localVmdkFile)
	if err != nil {
		return nil, errors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	logrus.WithFields(logrus.Fields{
		"name":           params.Name,
		"id":             params.Name,
		"size":           sizeMb,
		"datastore-path": vsphereImageDir,
	}).Infof("importing base vmdk for unikernel image")

	if err := c.ImportVmdk(localVmdkFile, vsphereImageDir); err != nil {
		return nil, errors.New("importing base boot.vmdk to vsphere datastore", err)
	}

	image := &types.Image{
		Id:             params.Name,
		Name:           params.Name,
		StageSpec:      params.RawImage.StageSpec,
		RunSpec:        params.RawImage.RunSpec,
		SizeMb:         sizeMb,
		Infrastructure: types.Infrastructure_VSPHERE,
		Created:        time.Now(),
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
