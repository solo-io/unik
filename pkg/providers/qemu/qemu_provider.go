package qemu

import (
	"os"
	"path/filepath"

	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
)

var QemuStateFile = os.Getenv("HOME") + "/.unik/qemu/state.json"
var qemuImagesDirectory = os.Getenv("HOME") + "/.unik/qemu/images/"
var qemuInstancesDirectory = os.Getenv("HOME") + "/.unik/qemu/instances/"
var qemuVolumesDirectory = os.Getenv("HOME") + "/.unik/qemu/volumes/"

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
