package photon

import "github.com/emc-advanced-dev/pkg/errors"

func (p *PhotonProvider) DeleteImage(id string, force bool) error {
	image, err := p.GetImage(id)
	if err != nil {
		return errors.New("Delete image", err)
	}

	task, err := p.client.Images.Delete(image.InfrastructureId)
	if err != nil {
		return errors.New("Delete image", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Delete image", err)
	}
	return nil
}
