package virtualbox

import (
	"os"
	"time"

	"path/filepath"

	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/solo-io/unik/pkg/os"
	"github.com/solo-io/unik/pkg/providers/common"
	"github.com/solo-io/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/solo-io/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxhttpclient"
)

func (p *VirtualboxProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	if _, err := p.GetInstance(params.Name); err == nil {
		return nil, errors.New("instance with name "+params.Name+" already exists. virtualbox provider requires unique names for instances", nil)
	}

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, params.MntPointsToVolumeIds); err != nil {
		return nil, errors.New("invalid mapping for volume", err)
	}

	instanceDir := getInstanceDir(params.Name)

	portsUsed := []int{}

	logrus.Debugf("using storage controller %s", image.RunSpec.StorageDriver)

	defer func() {
		if err != nil {
			if params.NoCleanup {
				logrus.Warnf("because --no-cleanup flag was provided, not cleaning up failed instance %s.2", params.Name)
				return
			}
			logrus.WithError(err).Errorf("error encountered, ensuring vm and disks are destroyed")
			virtualboxclient.PowerOffVm(params.Name)
			for _, portUsed := range portsUsed {
				virtualboxclient.DetachDisk(params.Name, portUsed, image.RunSpec.StorageDriver)
			}
			virtualboxclient.DestroyVm(params.Name)
			os.RemoveAll(instanceDir)
		}
	}()

	//if not set, use default
	if params.InstanceMemory <= 0 {
		params.InstanceMemory = image.RunSpec.DefaultInstanceMemory
	}

	logrus.Debugf("creating virtualbox vm")

	if err := virtualboxclient.CreateVm(params.Name, virtualboxInstancesDirectory(), params.InstanceMemory, p.config.AdapterName, p.config.VirtualboxAdapterType, image.RunSpec.StorageDriver); err != nil {
		return nil, errors.New("creating vm", err)
	}

	logrus.Debugf("copying source boot vmdk")
	instanceBootImage := filepath.Join(instanceDir, "boot.vmdk")
	if err := unikos.CopyFile(getImagePath(image.Name), instanceBootImage); err != nil {
		return nil, errors.New("copying base boot image", err)
	}
	if err := virtualboxclient.RefreshDiskUUID(instanceBootImage); err != nil {
		return nil, errors.New("refreshing disk uuid", err)
	}
	if err := virtualboxclient.AttachDisk(params.Name, instanceBootImage, 0, image.RunSpec.StorageDriver); err != nil {
		return nil, errors.New("attaching boot vol to instance", err)
	}

	for mntPoint, volumeId := range params.MntPointsToVolumeIds {
		volume, err := p.GetVolume(volumeId)
		if err != nil {
			return nil, errors.New("getting volume", err)
		}
		controllerPort, err := common.GetControllerPortForMnt(image, mntPoint)
		if err != nil {
			return nil, errors.New("getting controller port for mnt point", err)
		}
		if err := virtualboxclient.RefreshDiskUUID(getVolumePath(volume.Name)); err != nil {
			return nil, errors.New("refreshing disk uuid", err)
		}
		if err := virtualboxclient.AttachDisk(params.Name, getVolumePath(volume.Name), controllerPort, image.RunSpec.StorageDriver); err != nil {
			return nil, errors.New("attaching to vm", err)
		}
		portsUsed = append(portsUsed, controllerPort)
	}

	logrus.Debugf("setting instance id from mac address")
	vm, err := virtualboxclient.GetVm(params.Name)
	if err != nil {
		return nil, errors.New("retrieving created vm from vbox", err)
	}
	macAddr := vm.MACAddr
	instanceId := vm.UUID

	instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout)
	if err != nil {
		return nil, errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	logrus.Debugf("sending env to listener")
	if _, _, err := lxhttpclient.Post(instanceListenerIp+":3000", "/set_instance_env?mac_address="+macAddr, nil, params.Env); err != nil {
		return nil, errors.New("sending instance env to listener", err)
	}

	logrus.Debugf("powering on vm")
	if err := virtualboxclient.PowerOnVm(params.Name); err != nil {
		return nil, errors.New("powering on vm", err)
	}

	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Pending,
		IpAddress:      "",
		Infrastructure: types.Infrastructure_VIRTUALBOX,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return nil, errors.New("modifying instance map in state", err)
	}

	logrus.WithField("instance", instance).Infof("instance created successfully")

	return instance, nil
}
