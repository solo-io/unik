package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"path/filepath"
	"strconv"
)

func (p *VsphereProvider) DetachVolume(id string) error {
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
	vm, err := virtualboxclient.GetVm(instance.Id)
	if err != nil {
		return lxerrors.New("retreiving vm from virtualbox", err)
	}
	var controllerKey string
	for _, device := range vm.Devices {
		if filepath.Clean(device.DiskFile) == filepath.Clean(getVolumeDatastorePath(volume.Name)) {
			controllerKey = device.ControllerKey
		}
	}
	if controllerKey == "" {
		return lxerrors.New("could not find device attached to "+instance.Name+" that matches volume "+ getVolumeDatastorePath(volume.Name), nil)
	}

	controllerPort, err := strconv.Atoi(controllerKey)
	if err != nil {
		return lxerrors.New("could not convert "+controllerKey+" to int", err)
	}
	if err := p.getClient().DetachDisk(instance.Id, controllerPort); err != nil {
		return lxerrors.New("detaching disk from vm", err)
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
