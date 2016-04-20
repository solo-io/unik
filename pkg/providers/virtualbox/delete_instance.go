package virtualbox

import (
	"github.com/layer-x/layerx-commons/lxerrors"
	"strconv"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *VirtualboxProvider) DeleteInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	image, err := p.GetImage(instance.ImageId)
	if err != nil {
		return lxerrors.New("getting image for instance", err)
	}
	if err := p.StopInstance(instance.Id); err != nil {
		return lxerrors.New("powering off instance", err)
	}
	for _, deviceMapping := range image.DeviceMappings {
		if deviceMapping.MountPoint != "/" {
			controllerPort, err := strconv.Atoi(deviceMapping.DeviceName)
			if err != nil {
				return lxerrors.New("could not convert "+deviceMapping.DeviceName+" to int", err)
			}
			if err := virtualboxclient.DetachDisk(instance.Name, controllerPort); err != nil {
				return lxerrors.New("detaching volume from instance", err)
			} //TODO: do this in a place where we modify state (detach vol)
		}
	}
	if err := virtualboxclient.DestroyVm(instance.Name); err != nil {
		return lxerrors.New("destroying vm", err)
	}
	err = p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	})
	if err != nil {
		return lxerrors.New("modifying image map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return lxerrors.New("saving image map to state", err)
	}
	return nil
}
