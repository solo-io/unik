package virtualbox

import (
	"github.com/layer-x/layerx-commons/lxerrors"
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
		return lxerrors.New("powering off instance", err)
	}
	for controllerPort, deviceMapping := range image.DeviceMappings {
		if deviceMapping.MountPoint != "/" {
			if err := virtualboxclient.DetachDisk(instance.Name, controllerPort); err != nil {
				return lxerrors.New("detaching volume from instance", err)
			}
		}
	}
	if err := virtualboxclient.DestroyVm(instance.Name); err != nil {
		return lxerrors.New("destroying vm", err)
	}
	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		delete(instances, instance.Id)
		return nil
	}); err != nil {
		return lxerrors.New("modifying image map in state", err)
	}
	for _, volume := range volumesToDetach {
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
	}
	err = p.state.Save()
	if err != nil {
		return lxerrors.New("saving image map to state", err)
	}
	return nil
}
