package xen

import (
	"os"
	"path/filepath"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/xen/xenclient"
	"github.com/emc-advanced-dev/unik/pkg/state"
)

type XenProvider struct {
	state  state.State
	client *xenclient.XenClient
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

func NewXenProvider(config config.Xen) (*XenProvider, error) {

	os.MkdirAll(xenImagesDirectory(), 0777)
	os.MkdirAll(xenInstancesDirectory(), 0777)
	os.MkdirAll(xenVolumesDirectory(), 0777)

	p := &XenProvider{
		state: state.NewBasicState(XenStateFile()),
		client: &xenclient.XenClient{
			KernelPath: config.KernelPath,
			XenBridge:  config.XenBridge,
		},
	}

	if err := p.deployInstanceListener(); err != nil {
		return nil, errors.New("deploying xen instance listener", err)
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

func getInstanceDir(instanceName string) string {
	return filepath.Join(xenInstancesDirectory(), instanceName)
}

func getVolumePath(volumeName string) string {
	return filepath.Join(xenVolumesDirectory(), volumeName, "data.img")
}
