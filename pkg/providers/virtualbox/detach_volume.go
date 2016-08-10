package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"path/filepath"
	"strconv"
)

func (p *VirtualboxProvider) DetachVolume(id string) error {
	volume, err := p.GetVolume(id)
	if err != nil {
		return errors.New("retrieving volume "+id, err)
	}
	if volume.Attachment == "" {
		return errors.New("volume has no attachment", nil)
	}
	instanceId := volume.Attachment
	instance, err := p.GetInstance(instanceId)
	if err != nil {
		return errors.New("retrieving instance "+instanceId, err)
	}
	image, err := p.GetImage(instance.ImageId)
	if err != nil {
		return errors.New("retrieving image "+instance.ImageId, err)
	}
	vm, err := virtualboxclient.GetVm(instance.Id)
	if err != nil {
		return errors.New("retreiving vm from virtualbox", err)
	}
	var controllerKey string
	for _, device := range vm.Devices {
		if filepath.Clean(device.DiskFile) == filepath.Clean(getVolumePath(volume.Name)) {
			controllerKey = device.ControllerKey
		}
	}
	if controllerKey == "" {
		return errors.New("could not find device attached to "+instance.Name+" that matches volume "+getVolumePath(volume.Name), nil)
	}

	controllerPort, err := strconv.Atoi(controllerKey)
	if err != nil {
		return errors.New("could not convert "+controllerKey+" to int", err)
	}
	logrus.Debugf("using storage controller %s", image.RunSpec.StorageDriver)

	if err := virtualboxclient.DetachDisk(instance.Id, controllerPort, image.RunSpec.StorageDriver); err != nil {
		return errors.New("detaching disk from vm", err)
	}

	if err := p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volume, ok := volumes[volume.Id]
		if !ok {
			return errors.New("no record of "+volume.Id+" in the state", nil)
		}
		volume.Attachment = ""
		return nil
	}); err != nil {
		return errors.New("modifying volume map in state", err)
	}
	err = p.state.Save()
	if err != nil {
		return errors.New("saving modified volume map to state", err)
	}
	return nil
}
