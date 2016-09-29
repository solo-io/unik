package ukvm

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/util"
)

func (p *UkvmProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	if _, err := p.GetInstance(params.Name); err == nil {
		return nil, errors.New("instance with name "+params.Name+" already exists. ukvm provider requires unique names for instances", nil)
	}

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("getting image", err)
	}

	if image.RunSpec.Compiler != compilers.MIRAGE_OCAML_UKVM.String() {
		return nil, errors.New("ukvm only supports mirage / ukvm", nil)
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

	logrus.Debugf("creating ukvm vm")

	volImagesInOrder, err := p.getVolumeImages(volumeIdInOrder)
	if err != nil {
		return nil, errors.New("can't get volumes", err)
	}

	volArgs := volPathToUkvmArgs(volImagesInOrder)

	ukvmArgs := []string{}

	if p.config.Tap != "" {
		ukvmArgs = append(ukvmArgs, fmt.Sprintf("--net=%s", p.config.Tap))

	}

	ukvmArgs = append(ukvmArgs, volArgs...)
	ukvmArgs = append(ukvmArgs, getKernelPath(image.Name))
	cmd := exec.Command(getUkvmPath(image.Name), ukvmArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logrus.WithError(err).Warning("Can't get stdout for logs")
	}

	instanceLogName := getInstanceLogName(params.Name)

	go func() {
		f, err := os.Create(instanceLogName)
		if err != nil {
			logrus.WithError(err).Warning("Failed to create stdout log for instance " + params.Name)
		}
		defer f.Close()
		io.Copy(f, stdout)

	}()

	util.LogCommand(cmd, true)

	if err := cmd.Start(); err != nil {
		return nil, errors.New("can't start ukvm.", nil)
	}
	// close command resources
	go cmd.Wait()

	var instanceIp string

	instance := &types.Instance{
		Id:             fmt.Sprintf("%v", cmd.Process.Pid),
		Name:           params.Name,
		State:          types.InstanceState_Running,
		IpAddress:      instanceIp,
		Infrastructure: types.Infrastructure_UKVM,
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

func (p *UkvmProvider) getVolumeImages(volumeIdInOrder []string) ([]string, error) {

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

func volPathToUkvmArgs(volPaths []string) []string {
	var res []string
	for _, v := range volPaths {
		res = append(res, fmt.Sprintf("--disk=%s", v))
	}
	return res
}
