package photon

import (
	"net/url"
	"os"
	"path/filepath"

	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"github.com/vmware/photon-controller-go-sdk/photon"
)

type PhotonProvider struct {
	config    config.Photon
	state     state.State
	u         *url.URL
	client    *photon.Client
	projectId string
}

func PhotonStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "photon/state.json")

}
func photonImagesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "photon/images/")
}

func photonInstancesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "photon/instances/")
}

func photonVolumesDirectory() string {
	return filepath.Join(config.Internal.UnikHome, "photon/volumes/")
}

func NewPhotonProvider(config config.Photon) (*PhotonProvider, error) {

	os.MkdirAll(photonImagesDirectory(), 0755)
	os.MkdirAll(photonInstancesDirectory(), 0755)
	os.MkdirAll(photonVolumesDirectory(), 0755)

	p := &PhotonProvider{
		config: config,
		state:  state.NewBasicState(PhotonStateFile()),
	}

	return p, nil
}

func (p *PhotonProvider) WithState(state state.State) *PhotonProvider {
	p.state = state
	return p
}
