package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"github.com/emc-advanced-dev/unik/pkg/providers/vsphere/api"
	"net/url"
	"github.com/layer-x/layerx-commons/lxerrors"
	"strings"
)

const awsStateFile = "/var/unik/aws_state.json"

type VsphereProvider struct {
	config config.Vsphere `json:"Config"`
	State  state.State    `json:"State"`
	u      url.URL
}

func NewVsphereProvier(config config.Vsphere) (*VsphereProvider, error) {
	rawUrl := "https://" + config.VsphereUser + ":" + config.VspherePassword + "@" + strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(config.VsphereURL, "http://"), "https://"), "/sdk") + "/sdk"
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, lxerrors.New("parsing vsphere url", err)
	}
	return &VsphereProvider{
		config: config,
		State:  state.NewMemoryState(awsStateFile),
		u: u,
	}
}

func (p *VsphereProvider) getClient() *api.VsphereClient {
	return api.NewVsphereClient(p.u)
}