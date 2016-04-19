package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/pwnall/vbox"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"path/filepath"
)

var virtualboxStateFile = os.Getenv("HOME")+"/.unik/virtualbox/state.json"
var virtualboxImagesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/images/"
var virtualboxInstancesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/instances/"
var virtualboxVolumesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/volumes/"

const VboxUnikInstanceListener = "VboxUnikInstanceListener"

type VirtualboxProvider struct {
	config config.Virtualbox
	state  state.LocalStorageState
}

func NewVirtualboxProvider(config config.Virtualbox) (*VirtualboxProvider, error) {
	err := vbox.Init()
	if err != nil {
		return nil, lxerrors.New("initializing virtualbox client", err)
	}
	
	os.MkdirAll(virtualboxImagesDirectory, 0644)
	os.MkdirAll(virtualboxVolumesDirectory, 0644)

	return &VirtualboxProvider{
		config: config,
		state: state.NewLocalStorageState(virtualboxStateFile),
	}, nil
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