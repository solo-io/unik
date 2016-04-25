package common

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strings"
)

func GetVolume(p providers.Provider, nameOrIdPrefix string) (*types.Volume, error) {
	volumes, err := p.ListVolumes()
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
