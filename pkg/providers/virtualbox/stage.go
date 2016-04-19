package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"os"
	"github.com/layer-x/layerx-commons/lxerrors"
	"time"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
)

func (p *VirtualboxProvider) Stage(name string, rawImage *types.RawImage, force bool) (_ *types.Image, err error) {
	images, err := p.ListImages()
	if err != nil {
		return nil, lxerrors.New("retrieving image list for existing image", err)
	}
	for _, image := range images {
		if image.Name == name {
			if !force {
				return nil, lxerrors.New("an image already exists with name '"+name+"', try again with --force", nil)
			} else {
				logrus.WithField("image", image).Warnf("force: deleting previous image with name "+name)
				err = p.DeleteImage(image.Id, true)
				if err != nil {
					return nil, lxerrors.New("removing previously existing image", err)
				}
			}
		}
	}
	bootImagePath := getImagePath(name)

	defer func() {
		logrus.Infof("cleaninng up image %s", rawImage.LocalImagePath)
		os.Remove(rawImage.LocalImagePath)
		if err != nil {
			os.Remove(bootImagePath)
		}
	}()

	logrus.WithField("raw-image", rawImage).Infof("creating boot volume from raw image")
	if err := common.ConvertRawImage("vmdk", bootImagePath); err != nil {
		return nil, lxerrors.New("converting raw image to vmdk", err)
	}

	rawImageFile, err := os.Stat(rawImage.LocalImagePath)
	if err != nil {
		return nil, lxerrors.New("statting raw image file", err)
	}
	sizeMb := rawImageFile.Size() >> 20

	logrus.WithFields(logrus.Fields{
		"name": name,
		"id": name,
		"size": sizeMb,
	}).Infof("creating base vmdk for unikernel image")

	image := &types.Image{
		Id: name,
		Name: name,
		DeviceMappings: rawImage.DeviceMappings,
		SizeMb: sizeMb,
		Infrastructure: types.Infrastructure_VIRTUALBOX,
		Created: time.Now(),
	}

	err = p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[name] = image
		return nil
	})
	if err != nil {
		return nil, lxerrors.New("modifying image map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return nil, lxerrors.New("saving image map to state", err)
	}

	logrus.WithFields(logrus.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}
