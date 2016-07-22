package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io/ioutil"
	"os"
)

func (p *VirtualboxProvider) PullImage(params types.PullImagePararms) error {
	images, err := p.ListImages()
	if err != nil {
		return errors.New("retrieving image list for existing image", err)
	}
	for _, image := range images {
		if image.Name == params.ImageName {
			if !params.Force {
				return errors.New("an image already exists with name '"+params.ImageName+"', try again with --force", nil)
			} else {
				logrus.WithField("image", image).Warnf("force: deleting previous image with name " + params.ImageName)
				if err := p.DeleteImage(image.Id, true); err != nil {
					logrus.Warn(errors.New("failed removing previously existing image", err))
				}
			}
		}
	}

	tmpImage, err := ioutil.TempFile("", "tmp-pull-image-"+params.ImageName)
	if err != nil {
		return errors.New("creating tmp file", err)
	}
	defer os.RemoveAll(tmpImage.Name())
	image, err := common.PullImage(params.Config, params.ImageName, tmpImage)
	if err != nil {
		return errors.New("pulling image", err)
	}
	imagePath := getImagePath(image.Name)
	if err := os.Rename(tmpImage.Name(), imagePath); err != nil {
		return errors.New("renaming tmp image to "+imagePath, err)
	}

	if err := p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[image.Name] = image
		return nil
	}); err != nil {
		return errors.New("modifying image map in state", err)
	}
	if err := p.state.Save(); err != nil {
		return errors.New("saving image map to state", err)
	}
	logrus.Infof("image %v pulled successfully from %v", err)
	return nil
}
