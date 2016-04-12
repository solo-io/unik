package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/vsphere/api"
	"net/url"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strings"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/state"
)

var VsphereStateFile = os.Getenv("HOME")+"/.unik/vsphere/state.json"
var VsphereImagesDirectory = os.Getenv("HOME")+"/.unik/vsphere/images/"
var VsphereVolumesDirectory = os.Getenv("HOME")+"/.unik/vsphere/volumes/"

type VsphereProvider struct {
	config      config.Vsphere
	state       state.LocalStorageState
	u           url.URL
}

func NewVsphereProvier(config config.Vsphere) (*VsphereProvider, error) {
	rawUrl := "https://" + config.VsphereUser + ":" + config.VspherePassword + "@" + strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(config.VsphereURL, "http://"), "https://"), "/sdk") + "/sdk"
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, lxerrors.New("parsing vsphere url", err)
	}
	os.MkdirAll(VsphereImagesDirectory, 0644)
	os.MkdirAll(VsphereVolumesDirectory, 0644)

	return &VsphereProvider{
		config: config,
		state:  state.NewLocalStorageState(VsphereStateFile),
		u: u,
	}
}

func (p *VsphereProvider) getClient() *api.VsphereClient {
	return api.NewVsphereClient(p.u)
}