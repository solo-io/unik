package qemu

import "github.com/emc-advanced-dev/unik/pkg/types"

func (p *XenProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}

	var instances []*types.Instance
	for _, v := range p.state.GetInstances() {
		instances = append(instances, v)
	}

	return instances, nil
}
