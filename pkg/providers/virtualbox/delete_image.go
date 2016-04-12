package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
)

func (p *VirtualboxProvider) DeleteImage(id string, force bool) error {
	image, err := p.GetImage(id)
	if err != nil {
		return lxerrors.New("retrieving image", err)
	}
	instances, err := p.ListInstances()
	if err != nil {
		return lxerrors.New("retrieving list of instances", err)
	}
	for _, instance := range instances {
		if instance.ImageId == image.Id {
			if !force {
				return lxerrors.New("instance "+instance.Id+" found which uses image "+image.Id+"; try again with force=true", nil)
			} else {
				err = p.DeleteInstance(instance.Id)
				if err != nil {
					return lxerrors.New("failed to delete instance "+instance.Id+" which is using image "+image.Id, err)
				}
			}
		}
	}

	imagePath, ok := p.state.GetImagePaths()[image.Id]
	if !ok {
		return lxerrors.New("could not find image file path for image "+image.Id, nil)
	}

	err = os.Remove(imagePath)
	if err != nil {
		return lxerrors.New("deleing image file at " + imagePath, err)
	}

	err = p.state.ModifyImages(func(images map[string]*types.Image) error {
		delete(images, image.Id)
		return nil
	})
	if err != nil {
		return lxerrors.New("modifying image map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return lxerrors.New("saving modified image map to state", err)
	}
	return nil

	err = p.state.ModifyImagePaths(func(imagePaths map[string]string) error {
		delete(imagePath, image.Id)
		return nil
	})
	if err != nil {
		return lxerrors.New("modifying image paths map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return lxerrors.New("saving modified image paths map to state", err)
	}

	return nil
}
