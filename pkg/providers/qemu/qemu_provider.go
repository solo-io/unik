package qemu

import (
	"os"
	"path/filepath"

	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
)

type QemuProvider struct {
	config    config.Qemu
	state     state.State
}

func QemuStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "qemu/state.json")

}
func qemuImagesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "qemu/images/")
}

func qemuInstancesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "qemu/instances/")
}

func qemuVolumesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "qemu/volumes/")
}

func NewQemuProvider(config config.Qemu) (*QemuProvider, error) {

	os.MkdirAll(qemuImagesDirectory(), 0777)
	os.MkdirAll(qemuInstancesDirectory(), 0777)
	os.MkdirAll(qemuVolumesDirectory(), 0777)

	p := &QemuProvider{
		config: config,
		state:  state.NewBasicState(QemuStateFile()),
	}

	return p, nil
}

func (p *QemuProvider) WithState(state state.State) *QemuProvider {
	p.state = state
	return p
}

func getImagePath(imageName string) string {
	return filepath.Join(qemuImagesDirectory(), imageName, "boot.img")
}

func getKernelPath(imageName string) string {
	return filepath.Join(qemuImagesDirectory(), imageName, "program.bin")
}

func getCmdlinePath(imageName string) string {
	return filepath.Join(qemuImagesDirectory(), imageName, "cmdline")
}

func getVolumePath(volumeName string) string {
	return filepath.Join(qemuVolumesDirectory(), volumeName, "data.img")
}
