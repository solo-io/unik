package virtualbox

import (
	"github.com/Sirupsen/logrus"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/emc-advanced-dev/unik/pkg/types"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/layer-x/layerx-commons/lxerrors"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"os"
	"time"
)

func (p *VirtualboxProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	if _, err := p.GetInstance(params.Name); err == nil {
		return nil, lxerrors.New("instance with name "+ params.Name +" already exists. virtualbox provider requires unique names for instances", nil)
	}

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, lxerrors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, params.MntPointsToVolumeIds); err != nil {
		return nil, lxerrors.New("invalid mapping for volume", err)
	}

	instanceDir := getInstanceDir(params.Name)

	portsUsed := []int{}

	defer func() {
		if err != nil {
			logrus.WithError(err).Errorf("error encountered, ensuring vm and disks are destroyed")
			virtualboxclient.PowerOffVm(params.Name)
			for _, portUsed := range portsUsed {
				virtualboxclient.DetachDisk(params.Name, portUsed)
			}
			virtualboxclient.DestroyVm(params.Name)
			os.RemoveAll(instanceDir)
		}
	}()

	logrus.Debugf("creating virtualbox vm")

	if err := virtualboxclient.CreateVm(params.Name, virtualboxInstancesDirectory, p.config.AdapterName, p.config.VirtualboxAdapterType); err != nil {
		return nil, lxerrors.New("creating vm", err)
	}

	logrus.Debugf("copying source boot vmdk")
	instanceBootImage := instanceDir + "/boot.vmdk"
	if err := unikos.CopyFile(getImagePath(image.Name), instanceBootImage); err != nil {
		return nil, lxerrors.New("copying base boot image", err)
	}
	if err := virtualboxclient.AttachDisk(params.Name, instanceBootImage, 0); err != nil {
		return nil, lxerrors.New("attaching boot vol to instance", err)
	}

	for mntPoint, volumeId := range params.MntPointsToVolumeIds {
		volume, err := p.GetVolume(volumeId)
		if err != nil {
			return nil, lxerrors.New("getting volume", err)
		}
		controllerPort, err := common.GetControllerPortForMnt(image, mntPoint)
		if err != nil {
			return nil, lxerrors.New("getting controller port for mnt point", err)
		}
		if err := virtualboxclient.AttachDisk(params.Name, getVolumePath(volume.Name), controllerPort); err != nil {
			return nil, lxerrors.New("attaching disk to vm", err)
		}
		portsUsed = append(portsUsed, controllerPort)
	}

	logrus.Debugf("setting instance id from mac address")
	vm, err := virtualboxclient.GetVm(params.Name)
	if err != nil {
		return nil, lxerrors.New("retrieving created vm from vbox", err)
	}
	macAddr := vm.MACAddr
	instanceId := vm.UUID

	instanceListenerIp, err := virtualboxclient.GetVmIp(VboxUnikInstanceListener)
	if err != nil {
		return nil, lxerrors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	logrus.Debugf("sending env to listener")
	if _, _, err := lxhttpclient.Post(instanceListenerIp+":3000", "/set_instance_env?mac_address="+macAddr, nil, params.Env); err != nil {
		return nil, lxerrors.New("sending instance env to listener", err)
	}

	logrus.Debugf("powering on vm")
	if err := virtualboxclient.PowerOnVm(params.Name); err != nil {
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

	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Pending,
		IpAddress:      instanceIp,
		Infrastructure: types.Infrastructure_VIRTUALBOX,
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
