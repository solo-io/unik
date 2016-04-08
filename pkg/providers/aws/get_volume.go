package aws

import (
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strings"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) GetVolume(logger lxlog.Logger, nameOrIdPrefix string) (*types.Volume, error) {
	volumes, err := p.ListVolumes(logger)
	if err != nil {
		return nil, lxerrors.New("retrieving volume list", err)
	}
	for _, volume := range volumes {
		if strings.Contains(volume.Id, nameOrIdPrefix) || strings.Contains(volume.Name, nameOrIdPrefix) {
			return volume, nil
		}
	}
	return nil, lxerrors.New("volume with name or id containing '"+nameOrIdPrefix+"' not found", nil)
}
