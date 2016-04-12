package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/layer-x/layerx-commons/lxerrors"
)

func (p *VsphereProvider) GetInstanceLogs(id string) (string, error) {
	instance, err := p.GetInstance(id)
	if err != nil {
		return "", lxerrors.New("retrieving instance "+id, err)
	}
	return common.GetInstanceLogs(instance)
}
