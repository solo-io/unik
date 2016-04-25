package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"os"
	"time"
	vspheretypes "github.com/vmware/govmomi/vim25/types"
)

func (p *VsphereProvider) RunInstance(name, imageId string, mntPointsToVolumeIds map[string]string, env map[string]string) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": imageId,
		"mounts":   mntPointsToVolumeIds,
		"env":      env,
	}).Infof("running instance %s", name)

	if _, err := p.GetInstance(name); err == nil {
		return nil, lxerrors.New("instance with name "+name+" already exists. virtualbox provider requires unique names for instances", nil)
	}

	image, err := p.GetImage(imageId)
	if err != nil {
		return nil, lxerrors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, mntPointsToVolumeIds); err != nil {
		return nil, lxerrors.New("invalid mapping for volume", err)
	}

	instanceDir := getInstanceDatastoreDir(name)

	portsUsed := []int{}

	c := p.getClient()

	defer func() {
		if err != nil {
			logrus.WithError(err).Errorf("error encountered, ensuring vm and disks are destroyed")
			c.PowerOffVm(name)
			for _, portUsed := range portsUsed {
				c.DetachDisk(name, portUsed)
			}
			c.DestroyVm(name)
			os.RemoveAll(instanceDir)
		}
	}()

	logrus.Debugf("creating vsphere vm")

	if err := c.CreateVm(name, p.config.DefaultInstanceMemory); err != nil {
		return nil, lxerrors.New("creating vm", err)
	}
	
	vm, err := c.GetVm(name)
	if err != nil {
		return nil, lxerrors.New("failed to retrieve vm info after create", err)
	}
	if vm.Config == nil {
		return nil, lxerrors.New("vm has no config object", nil)
	}
	macAddr := ""
	if vm.Config != nil && vm.Config.Hardware.Device != nil {
		FindEthLoop:
		for _, device := range vm.Config.Hardware.Device {
			switch device.(type) {
			case *vspheretypes.VirtualE1000:
				eth := device.(*vspheretypes.VirtualE1000)
				macAddr = eth.MacAddress
				break FindEthLoop
			case *vspheretypes.VirtualE1000e:
				eth := device.(*vspheretypes.VirtualE1000e)
				macAddr = eth.MacAddress
				break FindEthLoop
			case *vspheretypes.VirtualPCNet32:
				eth := device.(*vspheretypes.VirtualPCNet32)
				macAddr = eth.MacAddress
				break FindEthLoop
			case *vspheretypes.VirtualSriovEthernetCard:
				eth := device.(*vspheretypes.VirtualSriovEthernetCard)
				macAddr = eth.MacAddress
				break FindEthLoop
			case *vspheretypes.VirtualVmxnet:
				eth := device.(*vspheretypes.VirtualVmxnet)
				macAddr = eth.MacAddress
				break FindEthLoop
			case *vspheretypes.VirtualVmxnet2:
				eth := device.(*vspheretypes.VirtualVmxnet2)
				macAddr = eth.MacAddress
				break FindEthLoop
			case *vspheretypes.VirtualVmxnet3:
				eth := device.(*vspheretypes.VirtualVmxnet3)
				macAddr = eth.MacAddress
				break FindEthLoop
			}
		}
	}
	if macAddr == "" {
		logrus.WithFields(logrus.Fields{"vm": vm}).Warnf("vm found, cannot identify mac addr")
		return nil, lxerrors.New("could not find mac addr on vm", nil)
	}

	logrus.Debugf("copying base boot vmdk to instance dir")
	instanceBootImagePath := instanceDir + "/boot.vmdk"
	if err := c.CopyVmdk(getImageDatastorePath(image.Name), instanceBootImagePath); err != nil {
		return nil, lxerrors.New("copying base boot image", err)
	}
	if err := c.AttachDisk(name, instanceBootImagePath, 0); err != nil {
		return nil, lxerrors.New("attaching boot vol to instance", err)
	}

	for mntPoint, volumeId := range mntPointsToVolumeIds {
		volume, err := p.GetVolume(volumeId)
		if err != nil {
			return nil, lxerrors.New("getting volume", err)
		}
		controllerPort, err := common.GetControllerPortForMnt(image, mntPoint)
		if err != nil {
			return nil, lxerrors.New("getting controller port for mnt point", err)
		}
		if err := c.AttachDisk(name, getVolumeDatastorePath(volume.Name), controllerPort); err != nil {
			return nil, lxerrors.New("attaching disk to vm", err)
		}
		portsUsed = append(portsUsed, controllerPort)
	}

	instanceListenerIp, err := c.GetVmIp(VsphereUnikInstanceListener)
	if err != nil {
		return nil, lxerrors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	logrus.Debugf("sending env to listener")
	if _, _, err := lxhttpclient.Post(instanceListenerIp+":3000", "/set_instance_env?mac_address="+macAddr, nil, env); err != nil {
		return nil, lxerrors.New("sending instance env to listener", err)
	}

	logrus.Debugf("powering on vm")
	if err := c.PowerOnVm(name); err != nil {
		return nil, lxerrors.New("powering on vm", err)
	}

	var instanceIp string

	if err := unikutil.Retry(30, time.Duration(2000*time.Millisecond), func() error {
		logrus.Debugf("getting instance ip")
		instanceIp, err = common.GetInstanceIp(instanceListenerIp, 3000, macAddr)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, lxerrors.New("failed to retrieve instance ip", err)
	}

	instanceId := vm.Config.InstanceUuid
	instance := &types.Instance{
		Id:             instanceId,
		Name:           name,
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
		return nil, lxerrors.New("modifying instance map in state", err)
	}
	if err := p.state.Save(); err != nil {
		return nil, lxerrors.New("saving instance volume map to state", err)
	}

	logrus.WithField("instance", instance).Infof("instance created successfully")

	return instance, nil
}
