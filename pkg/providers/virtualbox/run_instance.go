package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxerrors"
	"time"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"os"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/layer-x/layerx-commons/lxhttpclient"
)

func (p *VirtualboxProvider) RunInstance(name, imageId string, mntPointsToVolumeIds map[string]string, env map[string]string) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
	"image-id": imageId,
		"mounts": mntPointsToVolumeIds,
		"env": env,
	}).Infof("running instance %s", name)

	image, err := p.GetImage(imageId)
	if err != nil {
		return nil, lxerrors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, mntPointsToVolumeIds); err != nil {
		return nil, lxerrors.New("invalid mapping for volume", err)
	}

	instanceDir := getInstanceDir(name)

	portsUsed := []int{}

	defer func(){
		if err != nil {
			logrus.WithError(err).Errorf("error encountered, ensuring vm and disks are destroyed")
			virtualboxclient.PowerOffVm(name)
			for _, portUsed := range portsUsed {
				virtualboxclient.DetachDisk(name, portUsed)
			}
			virtualboxclient.DestroyVm(name)
			os.RemoveAll(instanceDir)
		}
	}()

	logrus.Debugf("creating virtualbox vm")

	if err := virtualboxclient.CreateVm(name, virtualboxInstancesDirectory, p.config.HostNetworkName); err != nil {
		return nil, lxerrors.New("creating vm", err)
	}

	logrus.Debugf("copying source boot vmdk")
	instanceBootImage := instanceDir+"/boot.vmdk"
	if err := unikos.CopyFile(getImagePath(image.Name), instanceBootImage); err != nil {
		return nil, lxerrors.New("copying base boot image", err)
	}
	if err := virtualboxclient.AttachDisk(name, instanceBootImage, 0); err != nil {
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
		if err := virtualboxclient.AttachDisk(name, getVolumePath(volume.Name), controllerPort); err != nil {
			return nil, lxerrors.New("attaching disk to vm", err)
		}
		portsUsed = append(portsUsed, controllerPort)
	}

	if err := virtualboxclient.PowerOnVm(name); err != nil {
		return nil, lxerrors.New("powering on vm", err)
	}

	instanceListenerIp, err := virtualboxclient.GetVmIp(VboxUnikInstanceListener)
	if err != nil {
		return nil, lxerrors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	var instanceId, instanceIp string

	for retries := 0; retries < 5 ;retries++ {
		if err := func() error {
			logrus.Debugf("getting vm list")
			vm, err := virtualboxclient.GetVm(name)
			if err != nil {
				return lxerrors.New("retrieving created vm from vbox", err)
			}
			instanceId = vm.MACAddr
			logrus.Debugf("waiting for instance to register to listener")
			instanceIp, err = common.GetInstanceIp(instanceListenerIp, 3000, instanceId)
			if err != nil {
				return lxerrors.New("getting ip for instance from instancelistener", err)
			}
			return nil
		}(); err != nil {
			if retries > 4 {
				return nil, lxerrors.New("failed to register instance", err)
			} else {
				logrus.WithError(err).Warnf("instance not registered yet")
				time.Sleep(3000 * time.Millisecond)
			}
		} else {
			break
		}
	}

	if _, _, err := lxhttpclient.Post(instanceIp, "/inject_env", nil, env); err != nil {
		return nil, lxerrors.New("injecting env vars into instance", err)
	}

	//must add instance to state before attaching volumes
	instance := &types.Instance{
		Id: instanceId,
		Name: name,
		State: types.InstanceState_Pending,
		IpAddress: instanceIp,
		Infrastructure: types.Infrastructure_VIRTUALBOX,
		ImageId: image.Id,
		Created: time.Now(),
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error{
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
