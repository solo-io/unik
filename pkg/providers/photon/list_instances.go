package photon

import (
	"github.com/emc-advanced-dev/pkg/errors"

	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *PhotonProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}

	var instances []*types.Instance
	for _, instance := range p.state.GetInstances() {

		vm, err := p.client.VMs.Get(instance.Id)
		if err != nil {
			return nil, errors.New("retrieving vm for instance id "+instance.Id, err)
		}

		// TODO: get ip..

		switch vm.State {
		case "STARTED":
			instance.State = types.InstanceState_Running
		case "CREATING":
			instance.State = types.InstanceState_Pending
		case "STOPPED":
			fallthrough
		case "SUSPENDED":
			fallthrough
		default:
			instance.State = types.InstanceState_Stopped
			break
		}
		err = p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
			instances[instance.Id] = instance
			return nil
		})
		if err != nil {
			return nil, errors.New("saving instance to state", err)
		}

		instances = append(instances, instance)
	}

	return instances, nil
}
