package common

import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"strings"
)

func GetImage(logger lxlog.Logger, p providers.Provider, nameOrIdPrefix string) (*types.Image, error) {
	images, err := p.ListImages(logger)
	if err != nil {
		return nil, lxerrors.New("retrieving image list", err)
	}
	for _, image := range images {
		if strings.Contains(image.Id, nameOrIdPrefix) || strings.Contains(image.Name, nameOrIdPrefix) {
			return image, nil
		}
	}
	return nil, lxerrors.New("image with name or id containing '"+nameOrIdPrefix+"' not found", nil)
}
