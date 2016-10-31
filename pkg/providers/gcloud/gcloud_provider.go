package gcloud

import (
	"path/filepath"

	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/storage/v1"
)

func GcloudStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "gcloud", "state.json")
}

type GcloudProvider struct {
	config     config.Gcloud
	state      state.State
	computeSvc *compute.Service
	storageSvc *storage.Service
}

func NewGcloudProvier(config config.Gcloud) (*GcloudProvider, error) {
	logrus.Infof("state file: %s", GcloudStateFile())

	// Use oauth2.NoContext if there isn't a good context to pass in.
	ctx := context.Background()

	client, err := google.DefaultClient(ctx, compute.ComputeScope)
	if err != nil {
		return nil, errors.New("failed to start default client", err)
	}
	computeService, err := compute.New(client)
	if err != nil {
		return nil, errors.New("failed to start compute client", err)
	}

	storageSevice, err := storage.New(client)
	if err != nil {
		return nil, errors.New("failed to start storage client", err)
	}

	return &GcloudProvider{
		config:     config,
		state:      state.NewBasicState(GcloudStateFile()),
		computeSvc: computeService,
		storageSvc: storageSevice,
	}, nil
}

func (p *GcloudProvider) WithState(state state.State) *GcloudProvider {
	p.state = state
	return p
}

func (p *GcloudProvider) compute() *compute.Service {
	return p.computeSvc
}

func (p *GcloudProvider) storage() *storage.Service {
	return p.storageSvc
}
