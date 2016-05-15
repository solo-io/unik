package vsphere

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/instance-listener/bindata"
	"io/ioutil"
	"path/filepath"
	"github.com/emc-advanced-dev/unik/pkg/compilers/rump"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"time"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
)

var timeout = time.Second * 10

func (p *VsphereProvider) deployInstanceListener() (err error) {
	logrus.Infof("checking if instance listener is alive...")
	if instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout); err == nil {
		logrus.Infof("instance listener is alive with IP %s", instanceListenerIp)
		return nil
	}
	logrus.Infof("cannot contact instance listener... cleaning up previous if it exists..")
	c := p.getClient()
	c.DestroyVm(VsphereUnikInstanceListener)
	logrus.Infof("compiling new instance listener")
	sourceDir, err := ioutil.TempDir(unikutil.UnikTmpDir(), "")
	if err != nil {
		return errors.New("creating temp dir for instance listener source", err)
	}
	defer os.RemoveAll(sourceDir)
	rawImage, err := compileInstanceListener(sourceDir)
	if err != nil {
		return errors.New("compiling instance listener source to unikernel", err)
	}
	logrus.Infof("staging new instance listener image")
	c.Rmdir(getImageDatastoreDir(VsphereUnikInstanceListener))
	params := types.StageImageParams{
		Name: VsphereUnikInstanceListener,
		RawImage: rawImage,
		Force: true,
	}
	image, err := p.Stage(params)
	if err != nil {
		return errors.New("building bootable vsphere image for instsance listener", err)
	}
	defer func(){
		if err != nil {
			p.DeleteImage(image.Id, true)
		}
	}()

	if err := p.runInstanceListener(image); err != nil {
		return errors.New("launching instance of instance listener", err)
	}
	return nil
}

func compileInstanceListener(sourceDir string) (*types.RawImage, error) {
	mainData, err := bindata.Asset("instance-listener/main.go")
	if err != nil {
		return nil, errors.New("reading binary data of instance listener main", err)
	}
	if err := ioutil.WriteFile(filepath.Join(sourceDir, "main.go"), mainData, 0644); err != nil {
		return nil, errors.New("copying contents of instance listener main.go", err)
	}

	params := types.CompileImageParams{
		SourcesDir: sourceDir,
		Args: "-prefix "+instanceListenerPrefix,
		MntPoints: []string{"/data"},
	}
	rumpGoCompiler := &rump.RumpCompiler{
		DockerImage: "projectunik/compilers-rump-go-hw-no-wrapper",
		CreateImage: rump.CreateImageVmware,
	}
	return rumpGoCompiler.CompileRawImage(params)
}

func (p *VsphereProvider) runInstanceListener(image *types.Image) (err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": image.Id,
	}).Infof("running instance of instance listener")

	imagePath, err := unikos.BuildEmptyDataVolume(10)
	if err != nil {
		return errors.New("failed creating raw data volume", err)
	}
	defer os.Remove(imagePath)

	instanceListenerData := "InstanceListenerData"
	params := types.CreateVolumeParams{
		Name: instanceListenerData,
		ImagePath: imagePath,
	}
	instanceListenerVol, err := p.CreateVolume(params)
	if err != nil {
		return errors.New("creating data vol for instance listener", err)
	}

	c := p.getClient()

	instanceDir := getInstanceDatastoreDir(VsphereUnikInstanceListener)
	defer func() {
		if err != nil {
			logrus.WithError(err).Warnf("error encountered, ensuring vm and disks are destroyed")
			c.PowerOffVm(VsphereUnikInstanceListener)
			c.DestroyVm(VsphereUnikInstanceListener)
			c.Rmdir(instanceDir)
			p.DeleteVolume(instanceListenerVol.Id, true)
			c.Rmdir(getVolumeDatastorePath(instanceListenerData))
		}
	}()

	logrus.Debugf("creating vsphere vm")

	if err := c.CreateVm(VsphereUnikInstanceListener, image.RunSpec.DefaultInstanceMemory, image.RunSpec.VsphereNetworkType); err != nil {
		return errors.New("creating vm", err)
	}

	logrus.Debugf("copying base boot vmdk to instance dir")
	instanceBootImagePath := instanceDir + "/boot.vmdk"
	if err := c.CopyVmdk(getImageDatastorePath(image.Name), instanceBootImagePath); err != nil {
		return errors.New("copying base boot image", err)
	}
	if err := c.AttachDisk(VsphereUnikInstanceListener, instanceBootImagePath, 0, image.RunSpec.StorageDriver); err != nil {
		return errors.New("attaching boot vol to instance", err)
	}

	controllerPort, err := common.GetControllerPortForMnt(image, "/data")
	if err != nil {
		return errors.New("getting controller port for mnt point", err)
	}
	logrus.Infof("attaching %s to %s on controller port %v", instanceListenerVol.Id, VsphereUnikInstanceListener, controllerPort)
	if err := c.AttachDisk(VsphereUnikInstanceListener, getVolumeDatastorePath(instanceListenerVol.Name), controllerPort, image.RunSpec.StorageDriver); err != nil {
		return errors.New("attaching disk to vm", err)
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
	if err := p.state.Save(); err != nil {
		return errors.New("saving instance volume map to state", err)
	}

	logrus.Debugf("powering on vm")
	if err := c.PowerOnVm(VsphereUnikInstanceListener); err != nil {
		return errors.New("powering on vm", err)
	}

	instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, time.Second * 30)
	if err != nil {
		return errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	vm, err := c.GetVm(VsphereUnikInstanceListener)
	if err != nil {
		return errors.New("getting vm info from vsphere", err)
	}

	instanceId := vm.Config.UUID
	instance := &types.Instance{
		Id:             instanceId,
		Name:           VsphereUnikInstanceListener,
		State:          types.InstanceState_Pending,
		IpAddress:      instanceListenerIp,
		Infrastructure: types.Infrastructure_VSPHERE,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return errors.New("modifying instance map in state", err)
	}
	if err := p.state.Save(); err != nil {
		return errors.New("saving instance volume map to state", err)
	}
	logrus.WithField("instance", instance).Infof("instance created successfully")

	return nil
}