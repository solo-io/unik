package openstack

import (
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
)

func (p *OpenstackProvider) DeleteImage(id string, force bool) error {
	image, err := p.GetImage(id)
	if err != nil {
		return errors.New("retrieving image", err)
	}

	// Delete instances of this image.
	instances, err := p.ListInstances()
	if err != nil {
		return errors.New("failed to retrieve list of instances", err)
	}
	for _, instance := range instances {
		if instance.ImageId == image.Id {
			if !force {
				return fmt.Errorf("instance '%s' found which uses image '%s'! Try again with --force.", instance.Id, image.Id)
			} else {
				err = p.DeleteInstance(instance.Id, true)
				if err != nil {
					return errors.New(fmt.Sprintf("failed to delete instance '%s' which uses image '%s'", instance.Id, image.Id), err)
				}
			}
		}
	}

	clientGlance, err := p.newClientGlance()
	if err != nil {
		return err
	}

	if err := deleteImage(clientGlance, image.Id); err != nil {
		return errors.New(fmt.Sprintf("failed to delete image '%s'", image.Id), err)
	}

	// Update state.
	if err := p.state.ModifyImages(func(imageList map[string]*types.Image) error {
		delete(imageList, image.Id)
		return nil
	}); err != nil {
		return errors.New("failed to modify image map in state", err)
	}
	return nil
}

// deleteImage deletes image from OpenStack.
func deleteImage(clientGlance *gophercloud.ServiceClient, imageId string) error {
	return images.Delete(clientGlance, imageId).Err
}
