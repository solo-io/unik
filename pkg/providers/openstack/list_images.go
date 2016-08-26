package openstack

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"github.com/rackspace/gophercloud/pagination"
)

func (p *OpenstackProvider) ListImages() ([]*types.Image, error) {
	// Return immediately if no image is managed by unik.
	managedImages := p.state.GetImages()
	if len(managedImages) < 1 {
		return []*types.Image{}, nil
	}

	clientGlance, err := p.newClientGlance()
	if err != nil {
		return nil, err
	}

	return fetchImages(clientGlance, managedImages)
}

func fetchImages(clientGlance *gophercloud.ServiceClient, managedImages map[string]*types.Image) ([]*types.Image, error) {
	result := []*types.Image{}

	pager := images.List(clientGlance, nil)
	pager.EachPage(func(page pagination.Page) (bool, error) {
		imageList, err := images.ExtractImages(page)
		if err != nil {
			return false, err
		}

		for _, i := range imageList {
			// Filter out images that unik is not aware of.
			image, ok := managedImages[i.ID]
			if !ok {
				continue
			}
			result = append(result, image)
		}

		return true, nil
	})

	return result, nil
}
