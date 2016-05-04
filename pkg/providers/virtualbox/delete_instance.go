package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/Sirupsen/logrus"
)

func (p *VirtualboxProvider) DeleteInstance(id string, force bool) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return errors.New("retrieving instance "+id, err)
	}
	if instance.State == types.InstanceState_Running {
		if force {
			if err := p.StopInstance(instance.Id); err != nil {
				return errors.New("stopping instance for deletion", err)
			}
		} else {
			return errors.New("instance "+instance.Id+" is still running. try again with --force or power off instance first", err)
		}
	}
	image, err := p.GetImage(instance.ImageId)
	if err != nil {
		return errors.New("getting image for instance", err)
	}
	volumesToDetach := []*types.Volume{}
	volumes, err := p.ListVolumes()
	if err != nil {
		return errors.New("getting volume list", err)
	}
	for _, volume := range volumes {
		if volume.Attachment == instance.Id {
			volumesToDetach = append(volumesToDetach, volume)
		}
	}

	for controllerPort, deviceMapping := range image.DeviceMappings {
		if deviceMapping.MountPoint != "/" {
			storageType := getStorageType(image.ExtraConfig)
			logrus.Debugf("using storage controller %s", virtualboxclient.SCSI_Storage)

			switch storageType {
			case virtualboxclient.SCSI_Storage:
				if err := virtualboxclient.DetachDiskSCSI(instance.Id, controllerPort); err != nil {
					return errors.New("detaching scsi volume from instance", err)
				}
			case virtualboxclient.SATA_Storage:
				if err := virtualboxclient.DetachDiskSATA(instance.Id, controllerPort); err != nil {
					return errors.New("detaching sata volume from instance", err)
				}
			default:
				return errors.New("unknown storage type: "+string(storageType), nil)
			}
		}
	}
	if err := virtualboxclient.DestroyVm(instance.Id); err != nil {
		return errors.New("destroying vm", err)
	}
	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	}); err != nil {
		return errors.New("modifying image map in state", err)
	}
	for _, volume := range volumesToDetach {
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
	}
	err = p.state.Save()
	if err != nil {
		return errors.New("saving image map to state", err)
	}
	return nil
}
