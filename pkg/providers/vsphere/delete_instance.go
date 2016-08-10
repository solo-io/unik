package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *VsphereProvider) DeleteInstance(id string, force bool) error {
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
			return errors.New("instance "+instance.Id+"is still running. try again with --force or power off instance first", err)
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
			logrus.Debugf("detaching volume: %v", volume)
			volumesToDetach = append(volumesToDetach, volume)
		}
	}

	c := p.getClient()
	for controllerPort, deviceMapping := range image.RunSpec.DeviceMappings {
		if deviceMapping.MountPoint != "/" {
			if err := c.DetachDisk(instance.Id, controllerPort, image.RunSpec.StorageDriver); err != nil {
				return errors.New("detaching volume from instance", err)
			}
		}
	}
	err = c.DestroyVm(instance.Name)
	if err != nil {
		return errors.New("failed to terminate instance "+instance.Id, err)
	}
	return p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	})
}
