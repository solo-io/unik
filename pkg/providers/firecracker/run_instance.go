package firecracker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	firecrackersdk "github.com/firecracker-microvm/firecracker-go-sdk"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/solo-io/unik/pkg/providers/common"
	"github.com/solo-io/unik/pkg/types"
)

func toblk(s string) firecrackersdk.BlockDevice {
	return firecrackersdk.BlockDevice{
		HostPath: s,
		Mode:     "rw",
	}
}

func (p *FirecrackerProvider) RunInstance(params types.RunInstanceParams) (_ *types.Instance, err error) {

	logrus.WithFields(logrus.Fields{
		"image-id": params.ImageId,
		"mounts":   params.MntPointsToVolumeIds,
		"env":      params.Env,
	}).Infof("running instance %s", params.Name)

	if _, err := p.GetInstance(params.Name); err == nil {
		return nil, errors.New("instance with name "+params.Name+" already exists. firecracker provider requires unique names for instances", nil)
	}

	image, err := p.GetImage(params.ImageId)
	if err != nil {
		return nil, errors.New("getting image", err)
	}

	if err := common.VerifyMntsInput(p, image, params.MntPointsToVolumeIds); err != nil {
		return nil, errors.New("invalid mapping for volume", err)
	}

	instanceId := params.Name
	instanceDir := getInstanceDir(instanceId)

	err = os.Mkdir(instanceDir, 0755)
	if err != nil {
		return nil, errors.New("can't create instance dir", err)
	}

	logs := filepath.Join(instanceDir, "logs.fifo")
	metrics := filepath.Join(instanceDir, "metrics.fifo")
	sock := filepath.Join(instanceDir, "firecracker.sock")

	if params.InstanceMemory == 0 {
		params.InstanceMemory = image.RunSpec.DefaultInstanceMemory
	}

	rootDrive := getImagePath(image.Name)

	volumeIdInOrder := make([]string, len(params.MntPointsToVolumeIds))

	for mntPoint, volumeId := range params.MntPointsToVolumeIds {

		controllerPort, err := common.GetControllerPortForMnt(image, mntPoint)
		if err != nil {
			return nil, err
		}
		volumeIdInOrder[controllerPort] = volumeId
	}

	volImagesInOrder, err := p.getVolumeImages(volumeIdInOrder)
	if err != nil {
		return nil, errors.New("can't get volumes", err)
	}

	fcCfg := firecrackersdk.Config{
		BinPath:          p.config.Binary,
		SocketPath:       sock,
		LogFifo:          logs,
		LogLevel:         "Debug",
		MetricsFifo:      metrics,
		KernelImagePath:  p.config.Kernel,
		KernelArgs:       "console=ttyS0 reboot=k panic=1 pci=off",
		RootDrive:        toblk(rootDrive),
		AdditionalDrives: volPathToBlockDevices(volImagesInOrder),
		// TODO: add these later. NetworkInterfaces: NICs,
		CPUCount:    1,
		CPUTemplate: firecrackersdk.CPUTemplate("C3"),
		HtEnabled:   false,
		MemInMiB:    int64(params.InstanceMemory),
		Debug:       true,
		Console:     p.config.Console,
	}

	logrus.Debugf("creating firecracker vm")

	m, err := firecrackersdk.NewMachine(fcCfg, firecrackersdk.WithLogger(logrus.NewEntry(logrus.New())))
	if err != nil {
		logrus.Errorf("Failed creating machine: %s", err)
		return nil, err
	}

	ctx := context.Background()
	vmmCtx, vmmCancel := context.WithCancel(ctx)
	exitchan, err := m.Init(ctx)
	if err != nil {
		logrus.Errorf("Firecracker Init returned error %s", err)
		return nil, err
	}

	go func() {
		<-exitchan
		vmmCancel()
	}()

	err = m.StartInstance(vmmCtx)
	if err != nil {
		return nil, errors.New("can't start firecracker - make sure it's in your path.", nil)
	}

	// todo: once we have network support we can set this up.
	var instanceIp string

	instance := &types.Instance{
		Id:             instanceId,
		Name:           params.Name,
		State:          types.InstanceState_Running,
		IpAddress:      instanceIp,
		Infrastructure: types.Infrastructure_FIRECRACKER,
		ImageId:        image.Id,
		Created:        time.Now(),
	}

	go func() {
		<-vmmCtx.Done()
		p.state.RemoveInstance(instance)
		os.RemoveAll(instanceDir)
	}()

	if err := p.state.ModifyInstances(func(instances map[string]*types.Instance) error {
		instances[instance.Id] = instance
		return nil
	}); err != nil {
		return nil, errors.New("modifying instance map in state", err)
	}

	logrus.WithField("instance", instance).Infof("instance created successfully")

	p.mapLock.Lock()
	p.runningMachines[instanceId] = m
	p.mapLock.Unlock()

	return instance, nil
}

func (p *FirecrackerProvider) getVolumeImages(volumeIdInOrder []string) ([]string, error) {

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

func volPathToBlockDevices(volPaths []string) []firecrackersdk.BlockDevice {
	var res []firecrackersdk.BlockDevice
	for _, v := range volPaths {
		res = append(res, toblk(v))
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
