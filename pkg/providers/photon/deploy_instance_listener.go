package photon

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/compilers/rump"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

var timeout = time.Second * 10

const PhotonUnikInstanceListener = "PhotonUnikInstanceListener"
const instanceListenerPrefix = "unik_photon"

func (p *PhotonProvider) DeployInstanceListener(config config.Photon) error {
	logrus.Infof("checking if instance listener is alive...")
	if instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, timeout); err == nil {
		logrus.Infof("instance listener is alive with IP %s", instanceListenerIp)
		return nil
	}
	logrus.Infof("cannot contact instance listener... cleaning up previous if it exists..")
	vms, err := p.client.Projects.GetVMs(config.ProjectId, &photon.VmGetOptions{
		Name: PhotonUnikInstanceListener,
	})
	if err != nil {
		return errors.New("getting photon vm list", err)
	}
	for _, vm := range vms.Items {
		if vm.Name == PhotonUnikInstanceListener {
			task, err := p.client.VMs.Stop(vm.ID)
			if err != nil {
				return errors.New("Stopping vm", err)
			}
			task, _ = p.waitForTaskSuccess(task)
			p.client.VMs.Delete(vm.ID)
			break
		}
	}
	logrus.Infof("compiling new instance listener")
	sourceDir, err := ioutil.TempDir("", "photon.instancelistener.")
	if err != nil {
		return errors.New("creating temp dir for instance listener source", err)
	}
	defer os.RemoveAll(sourceDir)
	rawImage, err := common.CompileInstanceListener(sourceDir, instanceListenerPrefix, "compilers-rump-go-hw-no-stub", rump.CreateImageVmware, false)
	if err != nil {
		return errors.New("compiling instance listener source to unikernel", err)
	}
	defer os.Remove(rawImage.LocalImagePath)
	logrus.Infof("staging new instance listener image")
	//delete old image if it exists
	if err := p.deleteOldImage(); err != nil {
		logrus.Warn("failed removing previous image", err)
	}

	params := types.StageImageParams{
		Name:     PhotonUnikInstanceListener,
		RawImage: rawImage,
		Force:    true,
	}
	image, err := p.Stage(params)
	if err != nil {
		return errors.New("building bootable virtualbox image for instsance listener", err)
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

func (p *PhotonProvider) deleteOldImage() error {
	if err := p.DeleteImage(PhotonUnikInstanceListener, true); err != nil {
		return nil
	}
	images, err := p.client.Images.GetAll()
	if err != nil {
		return errors.New("retrieving photon image list", err)
	}
	for _, image := range images.Items {
		if image.Name == PhotonUnikInstanceListener {
			task, err := p.client.Images.Delete(image.ID)
			if err != nil {
				return errors.New("Delete image", err)
			}
			_, err = p.waitForTaskSuccess(task)
			if err != nil {
				return errors.New("Delete image", err)
			}
		}
	}
	return errors.New("previous image not found", err)
}

func (p *PhotonProvider) runInstanceListener(image *types.Image) (err error) {
	vmflavor, err := p.getUnikFlavor("vm")
	if err != nil {
		return errors.New("can't get vm flavor", err)
	}

	diskflavor, err := p.getUnikFlavor("ephemeral-disk")
	if err != nil {
		return errors.New("can't get disk flavor", err)
	}

	disk := photon.AttachedDisk{
		Flavor:   diskflavor.Name,
		Kind:     "ephemeral-disk",
		Name:     "bootdisk-" + image.Id,
		BootDisk: true,
	}

	vmspec := &photon.VmCreateSpec{
		Flavor:        vmflavor.Name,
		SourceImageID: image.Id,
		Name:          PhotonUnikInstanceListener,
		Affinities:    nil,
		AttachedDisks: []photon.AttachedDisk{disk},
	}

	task, err := p.client.Projects.CreateVM(p.projectId, vmspec)

	if err != nil {
		return errors.New("Creating vm", err)
	}

	task, err = p.waitForTaskSuccess(task)

	if err != nil {
		return errors.New("Waiting for create vm", err)
	}

	instanceId := task.Entity.ID
	task, err = p.client.VMs.Start(instanceId)
	if err != nil {
		return errors.New("Starting vm", err)
	}

	task, err = p.waitForTaskSuccess(task)
	if err != nil {
		return errors.New("Starting vm", err)
	}

	instanceListenerIp, err := common.GetInstanceListenerIp(instanceListenerPrefix, time.Minute*5)
	if err != nil {
		return errors.New("failed to retrieve instance listener ip. is unik instance listener running?", err)
	}

	instance := &types.Instance{
		Id:             instanceId,
		Name:           PhotonUnikInstanceListener,
		State:          types.InstanceState_Running,
		IpAddress:      instanceListenerIp,
		Infrastructure: types.Infrastructure_PHOTON,
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
