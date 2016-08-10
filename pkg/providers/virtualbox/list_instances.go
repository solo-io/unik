package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"time"
)

func (p *VirtualboxProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}
	var instances []*types.Instance
	for _, instance := range p.state.GetInstances() {
		vm, err := virtualboxclient.GetVm(instance.Name)
		if err != nil {
			return nil, errors.New("retrieving vm for instance id "+instance.Name, err)
		}
		macAddr := vm.MACAddr

		instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout)
		if err != nil {
			return nil, errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
		}

		if err := unikutil.Retry(5, time.Duration(1000*time.Millisecond), func() error {
			logrus.Debugf("getting instance ip")
			if instance.Name == VboxUnikInstanceListener {
				instance.IpAddress = instanceListenerIp
			} else {
				instance.IpAddress, err = common.GetInstanceIp(instanceListenerIp, 3000, macAddr)
				if err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			logrus.Warnf("failed to retrieve ip for instance %s. instance may be running but has not responded to udp broadcast", instance.Id)
		}

		switch vm.Running {
		case true:
			instance.State = types.InstanceState_Running
			break
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
