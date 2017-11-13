package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/providers/common"
	"github.com/solo-io/unik/pkg/providers/vsphere/vsphereclient"
	"github.com/solo-io/unik/pkg/types"
	unikutil "github.com/solo-io/unik/pkg/util"
	"time"
)

func (p *VsphereProvider) syncState() error {
	if len(p.state.GetInstances()) < 1 {
		return nil
	}
	c := p.getClient()
	vms := []*vsphereclient.VirtualMachine{}
	for instanceId := range p.state.GetInstances() {
		vm, err := c.GetVmByUuid(instanceId)
		if err != nil {
			return errors.New("getting vm info for "+instanceId, err)
		}
		vms = append(vms, vm)
	}
	for _, vm := range vms {
		//we use mac address as the vm id
		macAddr := ""
		for _, device := range vm.Config.Hardware.Device {
			if len(device.MacAddress) > 0 {
				macAddr = device.MacAddress
				break
			}
		}
		if macAddr == "" {
			logrus.WithFields(logrus.Fields{"vm": vm}).Warnf("vm found, cannot identify mac addr")
			continue
		}

		instanceId := vm.Config.UUID
		instance, ok := p.state.GetInstances()[instanceId]
		if !ok {
			continue
		}

		switch vm.Summary.Runtime.PowerState {
		case "poweredOn":
			instance.State = types.InstanceState_Running
			break
		case "poweredOff":
		case "suspended":
			instance.State = types.InstanceState_Stopped
			break
		default:
			instance.State = types.InstanceState_Unknown
			break
		}

		var ipAddress string
		unikutil.Retry(3, time.Duration(500*time.Millisecond), func() error {
			if instance.Name == VsphereUnikInstanceListener {
				ipAddress = p.instanceListenerIp
			} else {
				var err error
				ipAddress, err = common.GetInstanceIp(p.instanceListenerIp, 3000, macAddr)
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
			if _, ok := instances[instance.Id]; ok {
				instances[instance.Id].IpAddress = ipAddress
				instances[instance.Id].State = instance.State
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}
