package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strconv"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"path/filepath"
)

func (p *VirtualboxProvider) DetachVolume(id string) error {
	volume, err := p.GetVolume(id)
	if err != nil {
		return lxerrors.New("retrieving volume "+id, err)
	}
	if volume.Attachment == "" {
		return lxerrors.New("volume has no attachment", nil)
	}
	instanceId := volume.Attachment
	instance, err := p.GetInstance(instanceId)
	if err != nil {
		return lxerrors.New("retrieving instance "+instanceId, err)
	}
	vm, err := virtualboxclient.GetVm(instance.Name)
	if err != nil {
		return lxerrors.New("retreiving vm from virtualbox", err)
	}
	var controllerKey string
	for _,  device := range vm.Devices {
		if filepath.Clean(device.DiskFile) == filepath.Clean(getVolumePath(volume.Name)) {
			controllerKey = device.ControllerKey
		}
	}
	if controllerKey == "" {
		return lxerrors.New("could not find device attached to "+instance.Name+" that matches volume "+getVolumePath(volume.Name), nil)
	}

	controllerPort, err := strconv.Atoi(controllerKey)
	if err != nil {
		return lxerrors.New("could not convert "+controllerKey+" to int", err)
	}
	if err := virtualboxclient.DetachDisk(volume.Name, controllerPort); err != nil {
		return lxerrors.New("attaching disk to vm", err)
	}
	if err := p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volume, ok := volumes[volume.Id]
		if !ok {
			return lxerrors.New("no record of "+volume.Id+" in the state", nil)
		}
		volume.Attachment = ""
		return nil
	}); err != nil {
		return lxerrors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return lxerrors.New("saving modified volume map to state", err)
	}
	return nil
}
