package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/virtualbox/virtualboxclient"
	"github.com/layer-x/layerx-commons/lxerrors"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/pwnall/vbox"
)

var virtualboxStateFile = os.Getenv("HOME")+"/.unik/virtualbox/state.json"
var virtualboxImagesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/images/"
var virtualboxVolumesDirectory = os.Getenv("HOME")+"/.unik/virtualbox/volumes/"

type VirtualboxProvider struct {
	config config.Vsphere
	state  common.LocalStorageState
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
		state: common.NewLocalStorageState(virtualboxStateFile),
	}, nil
}

func (p *VirtualboxProvider) getClient() *api.VirtualboxClient {
	return api.NewVirtualboxClient()
}