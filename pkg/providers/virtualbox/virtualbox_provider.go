package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
	"path/filepath"
	"strings"
)

var VirtualboxStateFile = os.Getenv("HOME") + "/.unik/virtualbox/state.json"
var virtualboxImagesDirectory = os.Getenv("HOME") + "/.unik/virtualbox/images/"
var virtualboxInstancesDirectory = os.Getenv("HOME") + "/.unik/virtualbox/instances/"
var virtualboxVolumesDirectory = os.Getenv("HOME") + "/.unik/virtualbox/volumes/"

const VboxUnikInstanceListener = "VboxUnikInstanceListener"

type VirtualboxProvider struct {
	config config.Virtualbox
	state  state.State
}

func NewVirtualboxProvider(config config.Virtualbox) (*VirtualboxProvider, error) {
	os.MkdirAll(virtualboxImagesDirectory, 0777)
	os.MkdirAll(virtualboxVolumesDirectory, 0777)

	p := &VirtualboxProvider{
		config: config,
		state:  state.NewBasicState(VirtualboxStateFile),
	}

	if err := p.DeployInstanceListener(config); err != nil && !strings.Contains(err.Error(), "already exists") {
		return nil, lxerrors.New("deploing virtualbox instance listener", err)
	}

	return p, nil
}

func (p *VirtualboxProvider) WithState(state state.State) *VirtualboxProvider {
	p.state = state
	return p
}

func getImagePath(imageName string) string {
	return filepath.Join(virtualboxImagesDirectory, imageName, "boot.vmdk")
}

func getInstanceDir(instanceName string) string {
	return filepath.Join(virtualboxInstancesDirectory, instanceName)
}

func getVolumePath(volumeName string) string {
	return filepath.Join(virtualboxVolumesDirectory, volumeName, "data.vmdk")
}
