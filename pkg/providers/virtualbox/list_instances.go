package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/emc-advanced-dev/pkg/errors"
	"time"
)

func (p *VirtualboxProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}
	vms, err := virtualboxclient.Vms()
	if err != nil {
		return nil, errors.New("getting vms from virtualbox", err)
	}
	instances := []*types.Instance{}
	for _, vm := range vms {
		macAddr := vm.MACAddr
		instanceId := vm.UUID
		instance, ok := p.state.GetInstances()[instanceId]
		if !ok {
			logrus.WithFields(logrus.Fields{"vm": vm, "instance-id": macAddr}).Warnf("vm found that does not belong to unik, ignoring")
			continue
		}

		instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout)
		if err != nil {
			return nil, errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
		}

		if err := unikutil.Retry(5, time.Duration(2000*time.Millisecond), func() error {
			logrus.Debugf("getting instance ip")
			instance.IpAddress, err = common.GetInstanceIp(instanceListenerIp, 3000, macAddr)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, errors.New("failed to retrieve instance ip", err)
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
