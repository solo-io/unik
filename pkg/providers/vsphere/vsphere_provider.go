package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/providers/vsphere/api"
	"net/url"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strings"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
)

var vsphereStateFile = os.Getenv("HOME")+"/.unik/vsphere/state.json"
var vsphereImagesDirectory = os.Getenv("HOME")+"/.unik/vsphere/images/"
var vsphereVolumesDirectory = os.Getenv("HOME")+"/.unik/vsphere/volumes/"

type VsphereProvider struct {
	config      config.Vsphere
	state       common.LocalStorageState
	u           url.URL
}

func NewVsphereProvier(config config.Vsphere) (*VsphereProvider, error) {
	rawUrl := "https://" + config.VsphereUser + ":" + config.VspherePassword + "@" + strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(config.VsphereURL, "http://"), "https://"), "/sdk") + "/sdk"
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, lxerrors.New("parsing vsphere url", err)
	}
	os.MkdirAll(vsphereImagesDirectory, 0644)
	os.MkdirAll(vsphereVolumesDirectory, 0644)

	return &VsphereProvider{
		config: config,
		state:  common.NewLocalStorageState(vsphereStateFile),
		u: u,
	}
}

func (p *VsphereProvider) getClient() *api.VsphereClient {
	return api.NewVsphereClient(p.u)
}