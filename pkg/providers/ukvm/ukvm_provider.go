package ukvm

import (
	"os"
	"path/filepath"

	"github.com/cf-unik/unik/pkg/config"
	"github.com/cf-unik/unik/pkg/state"
)

type UkvmProvider struct {
	config config.Ukvm
	state  state.State
}

func UkvmStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "ukvm/state.json")

}
func ukvmImagesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "ukvm/images/")
}

func ukvmInstancesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "ukvm/instances/")
}

func ukvmVolumesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "ukvm/volumes/")
}

func NewUkvmProvider(config config.Ukvm) (*UkvmProvider, error) {

	os.MkdirAll(ukvmImagesDirectory(), 0777)
	os.MkdirAll(ukvmInstancesDirectory(), 0777)
	os.MkdirAll(ukvmVolumesDirectory(), 0777)

	p := &UkvmProvider{
		config: config,
		state:  state.NewBasicState(UkvmStateFile()),
	}

	return p, nil
}

func (p *UkvmProvider) WithState(state state.State) *UkvmProvider {
	p.state = state
	return p
}
func getImageDir(imageName string) string {
	return filepath.Join(ukvmImagesDirectory(), imageName)
}
func getKernelPath(imageName string) string {
	return filepath.Join(ukvmImagesDirectory(), imageName, "program.bin")
}
func getUkvmPath(imageName string) string {
	return filepath.Join(ukvmImagesDirectory(), imageName, "ukvm-bin")
}

func getInstanceDir(instanceName string) string {
	return filepath.Join(ukvmInstancesDirectory(), instanceName)
}

func getInstanceLogName(instanceName string) string {
	return filepath.Join(ukvmInstancesDirectory(), instanceName, "stdout")
}

func getVolumePath(volumeName string) string {
	return filepath.Join(ukvmVolumesDirectory(), volumeName, "data.img")
}
