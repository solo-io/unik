package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
	"path/filepath"
	"strconv"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
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
	storageType := getStorageType(instance.ExtraConfig)
	logrus.Debugf("using storage controller %s", storageType)

	switch storageType {
	case compilers.SCSI_Storage:
		if err := virtualboxclient.DetachDiskSCSI(instance.Id, controllerPort); err != nil {
			return errors.New("detaching scsi disk from vm", err)
		}
	case compilers.SATA_Storage:
		if err := virtualboxclient.DetachDiskSATA(instance.Id, controllerPort); err != nil {
			return errors.New("detaching sata disk from vm", err)
		}
	default:
		return errors.New("unknown storage type: "+string(storageType), nil)
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
