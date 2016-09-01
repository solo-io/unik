package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"strings"
	"time"
)

func (p *VirtualboxProvider) syncState() error {
	if len(p.state.GetInstances()) < 1 {
		return nil
	}
	for _, instance := range p.state.GetInstances() {
		vm, err := virtualboxclient.GetVm(instance.Name)
		if err != nil {
			if strings.Contains(err.Error(), "Could not find a registered machine") {
				logrus.Warnf("instance found in state that is no longer registered to Virtualbox")
				p.deleteInstanceFromState(instance)
				continue
			}
			return errors.New("retrieving vm for instance id "+instance.Name, err)
		}
		macAddr := vm.MACAddr

		if vm.Running {
			instance.State = types.InstanceState_Running
		} else {
			instance.State = types.InstanceState_Stopped
		}

		var ipAddress string
		if err := unikutil.Retry(3, time.Duration(500*time.Millisecond), func() error {
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
	}
	return nil
}
