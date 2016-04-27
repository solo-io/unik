package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/Sirupsen/logrus"
)

func (p *VsphereProvider) DeleteInstance(id string) error {
	instance, err := p.GetInstance(id)
	if err != nil {
		return lxerrors.New("retrieving instance "+id, err)
	}
	image, err := p.GetImage(instance.ImageId)
	if err != nil {
		return lxerrors.New("getting image for instance", err)
	}
	volumesToDetach := []*types.Volume{}
	volumes, err := p.ListVolumes()
	if err != nil {
		return lxerrors.New("getting volume list", err)
	}
	for _, volume := range volumes {
		if volume.Attachment == instance.Id {
			volumesToDetach = append(volumesToDetach, volume)
		}
	}
	if err := p.StopInstance(instance.Id); err != nil {
		logrus.WithError(err).Warnf("could not power off instance, is instance already powered off?")
	}

	c := p.getClient()
	for controllerPort, deviceMapping := range image.DeviceMappings {
		if deviceMapping.MountPoint != "/" {
			if err := c.DetachDisk(instance.Id, controllerPort); err != nil {
				return lxerrors.New("detaching volume from instance", err)
			}
		}
	}
	err = c.DestroyVm(instance.Name)
	if err != nil {
		return lxerrors.New("failed to terminate instance "+instance.Id, err)
	}
	return p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	})
}
