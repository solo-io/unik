package qemu

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"io/ioutil"
)

func (p *QemuProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	if _, err := p.GetInstance(params.Name); err == nil {
		return nil, errors.New("instance with name "+params.Name+" already exists. qemu provider requires unique names for instances", nil)
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

	logrus.Debugf("creating qemu vm")

	volImagesInOrder, err := p.getVolumeImages(volumeIdInOrder)
	if err != nil {
		return nil, errors.New("cant get volumes", err)
	}

	volArgs := volPathToQemuArgs(volImagesInOrder)

	cmdlinedata, err := ioutil.ReadFile(getCmdlinePath(image.Name))
	if err != nil {
		return nil, errors.New("reading cmdline file", err)
	}
	cmdline := injectEnv(string(cmdlinedata), params.Env)
	cmdline = strings.Replace(cmdline, ",", ",,", -1)

	qemuArgs := []string{"-m", fmt.Sprintf("%v", params.InstanceMemory), "-net",
		"nic,model=virtio,netdev=mynet0", "-netdev", "user,id=mynet0,net=192.168.76.0/24,dhcpstart=192.168.76.9",
		"-device", "virtio-blk-pci,id=blk0,drive=hd0",
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=none,id=hd0", getImagePath(image.Name)),
		"-kernel", getKernelPath(image.Name),
		"-append", cmdline,
	}

	if params.DebugMode {
		logrus.Debugf("running instance in debug mode.\nattach debugger with `unik debug --instance %s", params.Name)
		qemuArgs = append(qemuArgs, "-s", "-S")
	}

	if p.config.NoGraphic {
		qemuArgs = append(qemuArgs, "-nographic", "-vga", "none")
	}

	qemuArgs = append(qemuArgs, volArgs...)
	cmd := exec.Command("qemu-system-x86_64", qemuArgs...)

	util.LogCommand(cmd, true)

	if err := cmd.Start(); err != nil {
		return nil, errors.New("Can't start qemu - make sure it's in your path.", nil)
	}

	var instanceIp string

	instance := &types.Instance{
		Id:           params.Name,
		Name:           params.Name,
		State:          types.InstanceState_Running,
		IpAddress:      instanceIp,
		Infrastructure: types.Infrastructure_QEMU,
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

func (p *QemuProvider) getVolumeImages(volumeIdInOrder []string) ([]string, error) {

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

func volPathToQemuArgs(volPaths []string) []string {
	var res []string
	for _, v := range volPaths {
		res = append(res, "-drive", fmt.Sprintf("if=virtio,file=%s,format=qcow2", v))
	}
	return res
}

func injectEnv(cmdline string, env map[string]string) string {
	// rump json is not really json so we can't parse it
	var envRumpJson []string
	for key, value := range env {
		envRumpJson = append(envRumpJson, fmt.Sprintf("\"env\": \"%s=%s\"", key, value))
	}

	cmdline = cmdline[:len(cmdline)-2] + "," + strings.Join(envRumpJson, ",") + "}}"
	return cmdline
}
