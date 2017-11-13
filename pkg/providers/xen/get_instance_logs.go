package xen

import (
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/providers/common"
)

func (p *XenProvider) GetInstanceLogs(id string) (string, error) {
	instance, err := p.GetInstance(id)
	if err != nil {
		return "", errors.New("retrieving instance "+id, err)
	}
	return common.GetInstanceLogs(instance)
}
