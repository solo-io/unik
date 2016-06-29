package photon

import (
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/emc-advanced-dev/pkg/errors"

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

	p.client = photon.NewClient(p.config.PhotonURL, nil, nil)
	p.projectId = p.config.ProjectId
	_, err := p.client.Status.Get()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *PhotonProvider) WithState(state state.State) *PhotonProvider {
	p.state = state
	return p
}

func (p *PhotonProvider) waitForTaskSuccess(task *photon.Task) (*photon.Task, error) {
	task, err := p.client.Tasks.WaitTimeout(task.ID, 30*time.Minute)
	if err != nil {
		return nil, errors.New("error waiting for task creating photon image", err)
	}

	if task.State != "COMPLETED" {
		return nil, errors.New("Error with task "+task.ID, nil)
	}

	return task, nil
}
