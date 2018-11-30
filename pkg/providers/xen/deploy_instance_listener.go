package xen

import (
	"io/ioutil"
	"os"
	"time"

	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/compilers/rump"
	"github.com/solo-io/unik/pkg/providers/common"
	"github.com/solo-io/unik/pkg/providers/xen/xenclient"
	"github.com/solo-io/unik/pkg/types"
	"github.com/solo-io/unik/pkg/util"
)

const (
	instanceListenerPrefix  = "unik_xen"
	XenUnikInstanceListener = "XenUnikInstanceListener"
)

var timeout = time.Second * 10
var instanceListenerData = "InstanceListenerData"

func (p *XenProvider) deployInstanceListener() error {
	logrus.Infof("checking if instance listener is alive...")
	if instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout); err == nil {
		logrus.Infof("instance listener is alive with IP %s", instanceListenerIp)
		return nil
	}
	logrus.Infof("cannot contact instance listener... cleaning up previous if it exists..")
	p.client.DestroyVm(XenUnikInstanceListener)
	logrus.Infof("compiling new instance listener")
	sourceDir, err := ioutil.TempDir("", "xen.instancelistener.")
	if err != nil {
		return errors.New("creating temp dir for instance listener source", err)
	}
	defer os.RemoveAll(sourceDir)
	rawImage, err := common.CompileInstanceListener(sourceDir, instanceListenerPrefix, "compilers-rump-go-xen", rump.CreateImageXen, false)
	if err != nil {
		return errors.New("compiling instance listener source to unikernel", err)
	}
	defer os.Remove(rawImage.LocalImagePath)
	logrus.Infof("staging new instance listener image")
	os.RemoveAll(getImagePath(XenUnikInstanceListener))
	params := types.StageImageParams{
		Name:     XenUnikInstanceListener,
		RawImage: rawImage,
		Force:    true,
	}
	image, err := p.Stage(params)
	if err != nil {
		return errors.New("building bootable xen image for instsance listener", err)
	}
	defer func() {
		if err != nil {
			p.DeleteImage(image.Id, true)
		}
	}()

	if err := p.runInstanceListener(image); err != nil {
		return errors.New("launching instance of instance listener", err)
	}
	return nil
}

func (p *XenProvider) runInstanceListener(image *types.Image) (err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": image.Id,
	}).Infof("running instance of instance listener")

	newVolume := false
	instanceListenerVol, err := p.GetVolume(instanceListenerData)
	if err != nil {
		newVolume = true
		imagePath, err := util.BuildEmptyDataVolume(10)
		if err != nil {
			return errors.New("failed creating raw data volume", err)
		}
		defer os.RemoveAll(imagePath)
		createVolumeParams := types.CreateVolumeParams{
			Name:      instanceListenerData,
			ImagePath: imagePath,
		}

		instanceListenerVol, err = p.CreateVolume(createVolumeParams)
		if err != nil {
			return errors.New("creating data vol for instance listener", err)
		}
		defer func() {
			if err != nil {
				p.DeleteVolume(instanceListenerVol.Id, true)
			}
		}()
	}

	defer func() {
		if err != nil {
			logrus.WithError(err).Warnf("error encountered, ensuring vm and disks are destroyed")
			p.DetachVolume(instanceListenerVol.Id)
			p.client.DestroyVm(XenUnikInstanceListener)
			os.RemoveAll(getInstanceDir(XenUnikInstanceListener))
			if newVolume {
				os.RemoveAll(getVolumePath(instanceListenerData))
			}
		}
	}()

	logrus.Debugf("creating xen vm")

	xenParams := xenclient.CreateVmParams{
		Name:      XenUnikInstanceListener,
		Memory:    image.RunSpec.DefaultInstanceMemory,
		BootImage: getImagePath(image.Name),
		VmDir:     getInstanceDir(XenUnikInstanceListener),
		DataVolumes: []xenclient.VolumeConfig{
			xenclient.VolumeConfig{
				ImagePath:  getVolumePath(instanceListenerVol.Name),
				DeviceName: "sdb1",
			},
		},
	}

	os.MkdirAll(getInstanceDir(XenUnikInstanceListener), 0755)

	if err := p.client.CreateVm(xenParams); err != nil {
		return errors.New("creating vm", err)
	}

	if err := p.state.ModifyVolumes(func(volumes map[string]*types.Volume) error {
		volume, ok := volumes[instanceListenerVol.Id]
		if !ok {
			return errors.New("no record of "+volume.Id+" in the state", nil)
		}
		volume.Attachment = instanceListenerVol.Id
		return nil
	}); err != nil {
		return errors.New("modifying volumes in state", err)
	}

	instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, time.Minute*5)
	if err != nil {
		return errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	doms, err := p.client.ListVms()
	if err != nil {
		return errors.New("getting vm info from xen", err)
	}
	instanceId := XenUnikInstanceListener
	for _, d := range doms {
		if d.Config.CInfo.Name == XenUnikInstanceListener {
			instanceId = fmt.Sprintf("%d", d.Domid)
			break
		}
	}

	instance := &types.Instance{
		Id:             instanceId,
		Name:           XenUnikInstanceListener,
		State:          types.InstanceState_Pending,
		IpAddress:      instanceListenerIp,
		Infrastructure: types.Infrastructure_VIRTUALBOX,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return errors.New("modifying instance map in state", err)
	}
	logrus.WithField("instance", instance).Infof("instance created successfully")

	return nil
}
