package qemu

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
)

func (p *QemuProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {
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

	if len(params.MntPointsToVolumeIds) >= 1 {
		return nil, errors.New("qemu doesn't support volumes currently.", nil)
	}

	instanceDir := getInstanceDir(params.Name)
	os.Mkdir(instanceDir, 0755)

	defer func() {
		if err != nil {
			if params.NoCleanup {
				logrus.Warnf("because --no-cleanup flag was provided, not cleaning up failed instance %s.2", params.Name)
				return
			}
			logrus.WithError(err).Errorf("error encountered, ensuring vm and disks are destroyed")
			os.RemoveAll(instanceDir)
		}
	}()

	logrus.Debugf("creating qemu vm")

	// unzip the image
	cmdline, err := unzipImage(getImagePath(image.Name), instanceDir)
	if err != nil {
		return nil, errors.New("unzipping image", err)
	}
	kernel := getKernelFileName(instanceDir)
	// TODO run qemu

	// qemu double comma escape
	cmdline = strings.Replace(cmdline, ",", ",,", -1)

	cmd := exec.Command("qemu-system-x86_64", "-m", "128", "-net", "nic,model=virtio", "-kernel", kernel, "-append", cmdline)

	if err := cmd.Start(); err != nil {
		return nil, errors.New("Can't start qemu - make sure it's in your path.", nil)
	}

	instanceId := fmt.Sprintf("%d", cmd.Process.Pid)

	var instanceIp string

	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Pending,
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

const kernelFileName = "kernel"

func getKernelFileName(instanceDir string) string {
	return path.Join(instanceDir, kernelFileName)
}

func unzipImage(imagezip, instanceDir string) (string, error) {

	r, err := zip.OpenReader(imagezip)
	if err != nil {
		return "", err
	}

	kernelFile, err := os.Create(getKernelFileName(instanceDir))
	if err != nil {
		return "", err
	}

	var cmdlineBuf bytes.Buffer

	for _, f := range r.File {
		switch f.Name {
		case config.QemuKernelFileName:
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			_, err = io.Copy(kernelFile, rc)
			if err != nil {
				return "", err
			}

		case config.QemuArgsFileName:
			rc, err := f.Open()
			if err != nil {
				return "", err
			}
			_, err = io.Copy(&cmdlineBuf, rc)
			if err != nil {
				return "", err
			}
		default:
			return "", errors.New("unkown file in image", nil)
		}
	}

	cmdline := cmdlineBuf.String()
	return cmdline, nil
}
