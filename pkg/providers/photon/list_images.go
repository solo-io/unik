package photon

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *PhotonProvider) ListImages() ([]*types.Image, error) {
	images := []*types.Image{}
	for _, image := range p.state.GetImages() {
		images = append(images, image)
	}
	return images, nil
}
