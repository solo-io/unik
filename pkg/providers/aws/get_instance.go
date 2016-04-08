package aws

import (
	"github.com/layer-x/layerx-commons/lxlog"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strings"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *AwsProvider) GetInstance(logger lxlog.Logger, nameOrIdPrefix string) (*types.Instance, error) {
	instances, err := p.ListInstances(logger)
	if err != nil {
		return nil, lxerrors.New("retrieving instance list", err)
	}
	for _, instance := range instances {
		if strings.Contains(instance.Id, nameOrIdPrefix) || strings.Contains(instance.Name, nameOrIdPrefix) {
			return instance, nil
		}
	}
	return nil, lxerrors.New("instance with name or id containing '"+nameOrIdPrefix+"' not found", nil)
}
