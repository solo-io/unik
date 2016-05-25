package qemu

import (
	"os"
	"path/filepath"

	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
)

var QemuStateFile = filepath.Join(config.Internal.UnikHome, "qemu/state.json")
var qemuImagesDirectory = filepath.Join(config.Internal.UnikHome, "qemu/images/")
var qemuInstancesDirectory = filepath.Join(config.Internal.UnikHome, "qemu/instances/")
var qemuVolumesDirectory = filepath.Join(config.Internal.UnikHome, "qemu/volumes/")

type QemuProvider struct {
	config config.Qemu
	state  state.State
}

func NewQemuProvider(config config.Qemu) (*QemuProvider, error) {
	os.MkdirAll(qemuImagesDirectory, 0777)
	os.MkdirAll(qemuInstancesDirectory, 0777)
	os.MkdirAll(qemuVolumesDirectory, 0777)

	p := &QemuProvider{
		config: config,
		state:  state.NewBasicState(QemuStateFile),
	}

	return p, nil
}

func (p *QemuProvider) WithState(state state.State) *QemuProvider {
	p.state = state
	return p
}

func getImagePath(imageName string) string {
	return filepath.Join(qemuImagesDirectory, imageName, "boot.img")
}

func getInstanceDir(instanceName string) string {
	return filepath.Join(qemuInstancesDirectory, instanceName)
}

func getVolumePath(volumeName string) string {
	return filepath.Join(qemuVolumesDirectory, volumeName, "data.img")
}
