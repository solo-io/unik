package gcloud

import (
	"github.com/emc-advanced-dev/pkg/errors"
)

func (p *GcloudProvider) DeleteImage(id string, force bool) error {
	image, err := p.GetImage(id)
	if err != nil {
		return errors.New("retrieving image", err)
	}
	instances, err := p.ListInstances()
	if err != nil {
		return errors.New("retrieving list of instances", err)
	}
	for _, instance := range instances {
		if instance.ImageId == image.Id {
			if !force {
				return errors.New("instance "+instance.Id+" found which uses image "+image.Id+"; try again with force=true", nil)
			} else {
				err = p.DeleteInstance(instance.Id, true)
				if err != nil {
					return errors.New("failed to delete instance "+instance.Id+" which is using image "+image.Id, err)
				}
			}
		}
	}

	if _, err := p.compute().Images.Delete(p.config.ProjectID, image.Name).Do(); err != nil {
		return errors.New("deleting image from gcloud", err)
	}
	return p.state.RemoveImage(image)
}
