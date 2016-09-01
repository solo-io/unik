package virtualbox

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
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

	for controllerPort, deviceMapping := range image.RunSpec.DeviceMappings {
		if deviceMapping.MountPoint != "/" {
			logrus.Debugf("using storage controller %s", image.RunSpec.StorageDriver)
			if err := virtualboxclient.DetachDisk(instance.Id, controllerPort, image.RunSpec.StorageDriver); err != nil {
				return errors.New("detaching scsi volume from instance", err)
			}
		}
	}
	if err := virtualboxclient.DestroyVm(instance.Id); err != nil {
		return errors.New("destroying vm", err)
	}
	return p.state.RemoveInstance(instance)
}
