package photon

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *PhotonProvider) DeleteImage(id string, force bool) error {
	image, err := p.GetImage(id)
	if err != nil {
		return errors.New("image does not exist", err)
	}

	task, err := p.client.Images.Delete(image.Id)
	if err != nil {
		return errors.New("Delete image", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Delete image", err)
	}

	if err := p.state.ModifyImages(func(images map[string]*types.Image) error {
		delete(images, image.Id)
		return nil
	}); err != nil {
		return errors.New("modifying image map in state", err)
	}

	return nil
}
