package xen

import (
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/providers/xen/xenclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"os"
)

func (p *XenProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	if _, err := p.GetInstance(params.Name); err == nil {
		return nil, errors.New("instance with name "+params.Name+" already exists. xen provider requires unique names for instances", nil)
	}

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, params.MntPointsToVolumeIds); err != nil {
		return nil, errors.New("invalid mapping for volume", err)
	}

	volumeIdInOrder := make([]string, len(params.MntPointsToVolumeIds))

	for mntPoint, volumeId := range params.MntPointsToVolumeIds {
		controllerPort, err := common.GetControllerPortForMnt(image, mntPoint)
		if err != nil {
			return nil, err
		}
		volumeIdInOrder[controllerPort] = volumeId
	}

	logrus.Debugf("creating xen vm")

	volImagesInOrder, err := p.getVolumeImages(volumeIdInOrder)
	if err != nil {
		return nil, errors.New("can't get volumes", err)
	}

	dataVolumes := make([]xenclient.VolumeConfig, len(volImagesInOrder))
	for i, volPath := range volImagesInOrder {
		dataVolumes[i] = xenclient.VolumeConfig{
			ImagePath:  volPath,
			DeviceName: fmt.Sprintf("sd%c1", 'a'+i+1),
		}
	}

	if err := os.MkdirAll(getInstanceDir(params.Name), 0755); err != nil {
		return nil, errors.New("failed to create instance dir", err)
	}

	xenParams := xenclient.CreateVmParams{
		Name:        params.Name,
		Memory:      params.InstanceMemory,
		BootImage:   getImagePath(image.Name),
		VmDir:       getInstanceDir(params.Name),
		DataVolumes: dataVolumes,
	}

	if err := p.client.CreateVm(xenParams); err != nil {
		return nil, errors.New("creating xen domain", err)
	}

	instanceId := params.Name
	if doms, err := p.client.ListVms(); err == nil {
		for _, d := range doms {
			if d.Config.CInfo.Name == params.Name {
				instanceId = fmt.Sprintf("%d", d.Domid)
				break
			}
		}
	}

	var instanceIp string

	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Running,
		IpAddress:      instanceIp,
		Infrastructure: types.Infrastructure_XEN,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return nil, errors.New("modifying instance map in state", err)
	}
	if err := p.state.Save(); err != nil {
		return nil, errors.New("saving instance volume map to state", err)
	}

	logrus.WithField("instance", instance).Infof("instance created successfully")

	return instance, nil
}

func (p *XenProvider) getVolumeImages(volumeIdInOrder []string) ([]string, error) {
	var volPath []string
	for _, v := range volumeIdInOrder {
		v, err := p.GetVolume(v)
		if err != nil {
			return nil, err
		}
		volPath = append(volPath, getVolumePath(v.Name))
	}
	return volPath, nil
}
