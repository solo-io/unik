package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/config"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"path/filepath"
)

var VirtualboxStateFile = os.Getenv("HOME")+"/.unik/virtualbox/state.json"
var virtualboxImagesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/images/"
var virtualboxInstancesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/instances/"
var virtualboxVolumesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/volumes/"

const VboxUnikInstanceListener = "VboxUnikInstanceListener"

type VirtualboxProvider struct {
	config config.Virtualbox
	state  state.State
}

func NewVirtualboxProvider(config config.Virtualbox) *VirtualboxProvider {
	os.MkdirAll(virtualboxImagesDirectory, 0777)
	os.MkdirAll(virtualboxVolumesDirectory, 0777)

	return &VirtualboxProvider{
		config: config,
		state: state.NewBasicState(VirtualboxStateFile),
	}
}

func (p *VirtualboxProvider) WithState(state state.State) *VirtualboxProvider {
	p.state = state
	return p
}

func getImagePath(imageName string) string {
	return filepath.Join(virtualboxImagesDirectory, imageName,"boot.vmdk")
}

func getInstanceDir(instanceName string) string {
	return filepath.Join(virtualboxInstancesDirectory, instanceName)
}

func getVolumePath(volumeName string) string {
	return filepath.Join(virtualboxVolumesDirectory, volumeName, "boot.vmdk")
}