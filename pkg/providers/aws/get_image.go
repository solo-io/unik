package aws

import (
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strings"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) GetImage(logger lxlog.Logger, nameOrIdPrefix string) (*types.Image, error) {
	images, err := p.ListImages(logger)
	if err != nil {
		return nil, lxerrors.New("retrieving ec2 image list", err)
	}
	for _, image := range images {
		if strings.Contains(image.Id, nameOrIdPrefix) || strings.Contains(image.Name, nameOrIdPrefix) {
			return image, nil
		}
	}
	return nil, lxerrors.New("image with name or id containing '"+nameOrIdPrefix+"' not found", nil)
}
