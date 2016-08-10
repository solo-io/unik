package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"time"
)

func (p *VsphereProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
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

	instanceDir := getInstanceDatastoreDir(params.Name)

	portsUsed := []int{}

	c := p.getClient()

	defer func() {
		if err != nil {
			if params.NoCleanup {
				logrus.Warnf("because --no-cleanup flag was provided, not cleaning up failed instance %s001", params.Name)
				return
			}
			logrus.WithError(err).Warnf("error encountered, ensuring vm and disks are destroyed")
			c.PowerOffVm(params.Name)
			for _, portUsed := range portsUsed {
				c.DetachDisk(params.Name, portUsed, image.RunSpec.StorageDriver)
			}
			c.DestroyVm(params.Name)
			c.Rmdir(instanceDir)
		}
	}()

	logrus.Debugf("creating vsphere vm")

	//if not set, use default
	if params.InstanceMemory <= 0 {
		params.InstanceMemory = image.RunSpec.DefaultInstanceMemory
	}

	if err := c.CreateVm(params.Name, params.InstanceMemory, image.RunSpec.VsphereNetworkType, p.config.NetworkLabel); err != nil {
		return nil, errors.New("creating vm", err)
	}

	logrus.Debugf("powering on vm to assign mac addr")
	if err := c.PowerOnVm(params.Name); err != nil {
		return nil, errors.New("failed to power on vm to assign mac addr", err)
	}

	vm, err := c.GetVm(params.Name)
	if err != nil {
		return nil, errors.New("failed to retrieve vm info after create", err)
	}

	macAddr := ""
	if vm.Config.Hardware.Device != nil {
		for _, device := range vm.Config.Hardware.Device {
			if len(device.MacAddress) > 0 {
				macAddr = device.MacAddress
				break
			}
		}
	}
	if macAddr == "" {
		logrus.WithFields(logrus.Fields{"vm": vm}).Warnf("vm found, cannot identify mac addr")
		return nil, errors.New("could not find mac addr on vm", nil)
	}
	if err := c.PowerOffVm(params.Name); err != nil {
		return nil, errors.New("failed to power off vm after retrieving mac addr", err)
	}

	logrus.Debugf("copying base boot vmdk to instance dir")
	instanceBootImagePath := instanceDir + "/boot.vmdk"
	if err := c.CopyVmdk(getImageDatastorePath(image.Name), instanceBootImagePath); err != nil {
		return nil, errors.New("copying base boot image", err)
	}
	if err := c.AttachDisk(params.Name, instanceBootImagePath, 0, image.RunSpec.StorageDriver); err != nil {
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
		if err := c.AttachDisk(params.Name, getVolumeDatastorePath(volume.Name), controllerPort, image.RunSpec.StorageDriver); err != nil {
			return nil, errors.New("attaching disk to vm", err)
		}
		portsUsed = append(portsUsed, controllerPort)
	}

	instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout)
	if err != nil {
		return nil, errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	logrus.Debugf("sending env to listener")
	if _, _, err := lxhttpclient.Post(instanceListenerIp+":3000", "/set_instance_env?mac_address="+macAddr, nil, params.Env); err != nil {
		return nil, errors.New("sending instance env to listener", err)
	}

	logrus.Debugf("powering on vm")
	if err := c.PowerOnVm(params.Name); err != nil {
		return nil, errors.New("powering on vm", err)
	}

	var instanceIp string
	instanceId := vm.Config.UUID

	if err := unikutil.Retry(5, time.Duration(2000*time.Millisecond), func() error {
		logrus.Debugf("getting instance ip")
		instanceIp, err = common.GetInstanceIp(instanceListenerIp, 3000, macAddr)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		logrus.Warnf("failed to retrieve ip for instance %s. instance may be running but has not responded to udp broadcast", instanceId)
	}

	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Pending,
		IpAddress:      instanceIp,
		Infrastructure: types.Infrastructure_VSPHERE,
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
