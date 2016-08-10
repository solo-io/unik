package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/providers/vsphere/vsphereclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"time"
)

func (p *VsphereProvider) ListInstances() ([]*types.Instance, error) {
	if len(p.state.GetInstances()) < 1 {
		return []*types.Instance{}, nil
	}
	c := p.getClient()
	vms := []*vsphereclient.VirtualMachine{}
	for instanceId := range p.state.GetInstances() {
		vm, err := c.GetVmByUuid(instanceId)
		if err != nil {
			return nil, errors.New("getting vm info for "+instanceId, err)
		}
		vms = append(vms, vm)
	}
	instances := []*types.Instance{}
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

		instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout)
		if err != nil {
			return nil, errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
		}

		go func() {
			if err := unikutil.Retry(5, time.Duration(1000*time.Millisecond), func() error {
				logrus.Debugf("getting instance ip")
				if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
					if instance.Name == VsphereUnikInstanceListener {
						instances[instance.Id].IpAddress = instanceListenerIp
					} else {
						instances[instance.Id].IpAddress, err = common.GetInstanceIp(instanceListenerIp, 3000, macAddr)
						if err != nil {
							return err
						}
					}
					return nil
				}); err != nil {
					logrus.WithError(err).Warnf("failed to get instance " + instance.Name + " ip")
				}
				return nil
			}); err != nil {
				logrus.Warnf("failed to retrieve ip for instance %s. instance may be running but has not responded to udp broadcast", instance.Id)
			}
		}()

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
