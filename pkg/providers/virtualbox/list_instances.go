package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"time"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

func (p *VirtualboxProvider) ListInstances() ([]*types.Instance, error) {
	vms, err := virtualboxclient.Vms()
	if err != nil {
		return nil, lxerrors.New("getting vms from virtualbox", err)
	}
	instances := []*types.Instance{}
	for _, vm := range vms {
		instanceId := vm.MACAddr
		instance, ok := p.state.GetInstances()[instanceId]
		if !ok {
			logrus.WithFields(logrus.Fields{"vm": vm, "instance-id": instanceId}).Warnf("vm found, cannot identify instance id")
			continue
		}

		instanceListenerIp, err := virtualboxclient.GetVmIp(VboxUnikInstanceListener)
		if err != nil {
			return nil, lxerrors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
		}

		if err := unikutil.Retry(5, time.Duration(2000 * time.Millisecond), func() error {
			logrus.Debugf("getting instance ip")
			instance.IpAddress, err = common.GetInstanceIp(instanceListenerIp, 3000, instanceId)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return nil, lxerrors.New("failed to retrieve instance ip", err)
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
			return nil, lxerrors.New("saving instance to state", err)
		}

		instances = append(instances, instance)
	}
	return instances, nil
}
