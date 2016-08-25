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

		if vm.Running {
			instance.State = types.InstanceState_Running
		} else {
			instance.State = types.InstanceState_Stopped
		}

		go p.updateInstance(*instance, macAddr)

		instances = append(instances, instance)
	}
	return instances, nil
}

func (p *VirtualboxProvider) updateInstance(instance types.Instance, macAddr string) error {
	var ipAddress string
	if err := unikutil.Retry(5, time.Duration(1000*time.Millisecond), func() error {
		logrus.Debugf("getting instance ip")
		if instance.Name == VboxUnikInstanceListener {
			ipAddress = p.instanceListenerIp
		} else {
			var err error
			ipAddress, err = common.GetInstanceIp(p.instanceListenerIp, 3000, macAddr)
			if err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		logrus.Warnf("failed to retrieve ip for instance %s. instance may be running but has not responded to udp broadcast", instance.Id)
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		if _, ok := instances[instance.Id]; ok {
			instances[instance.Id].IpAddress = ipAddress
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
