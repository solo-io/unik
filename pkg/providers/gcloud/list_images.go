package gcloud

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *GcloudProvider) ListImages() ([]*types.Image, error) {
	images := []*types.Image{}
	for _, image := range p.state.GetImages() {
		if p.verifyImage(image.Name) {
			images = append(images, image)
		} else {
			p.state.ModifyImages(func(images map[string]*types.Image) error {
				delete(images, image.Id)
				return nil
			})
		}
	}
	return images, nil
}

func (p *GcloudProvider) verifyImage(imageName string) bool {
	_, err := p.compute().Images.Get(p.config.ProjectID, imageName).Do()
	return err == nil
}
