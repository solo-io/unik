package qemu

import (
	"os"
	"path/filepath"

	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"os/exec"
)

type XenProvider struct {
	config config.Xen
	state  state.State
}

func XenStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "xen/state.json")

}
func xenImagesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "xen/images/")
}

func xenInstancesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "xen/instances/")
}

func xenVolumesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "xen/volumes/")
}

func NewQemuProvider(config config.Qemu) (*XenProvider, error) {

	os.MkdirAll(xenImagesDirectory(), 0777)
	os.MkdirAll(xenInstancesDirectory(), 0777)
	os.MkdirAll(xenVolumesDirectory(), 0777)

	if config.DebuggerPort == 0 {
		config.DebuggerPort = 3001
	}

	p := &XenProvider{
		config: config,
		state:  state.NewBasicState(XenStateFile()),
	}

	return p, nil
}

func (p *XenProvider) WithState(state state.State) *XenProvider {
	p.state = state
	return p
}

func getImagePath(imageName string) string {
	return filepath.Join(xenImagesDirectory(), imageName, "boot.img")
}

func getKernelPath(imageName string) string {
	return filepath.Join(xenImagesDirectory(), imageName, "program.bin")
}

func getCmdlinePath(imageName string) string {
	return filepath.Join(xenImagesDirectory(), imageName, "cmdline")
}

func getVolumePath(volumeName string) string {
	return filepath.Join(xenVolumesDirectory(), volumeName, "data.img")
}
