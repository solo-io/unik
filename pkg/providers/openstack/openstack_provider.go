package openstack

import (
	"github.com/Sirupsen/logrus"
	"path/filepath"

	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/state"
	"github.com/rackspace/gophercloud"
)

type OpenstackProvider struct {
	config config.Openstack
	state  state.State
}

func OpenstackStateFile() string {
	return filepath.Join(config.Internal.UnikHome, "openstack/state.json")
}

func NewOpenstackProvider(config config.Openstack) (*OpenstackProvider, error) {
	logrus.Infof("openstack state file: %s", OpenstackStateFile())
	p := &OpenstackProvider{
		config: config,
		state:  state.NewBasicState(OpenstackStateFile()),
	}

	return p, nil
}

func (p *OpenstackProvider) WithState(state state.State) *OpenstackProvider {
	p.state = state
	return p
}

func (p *OpenstackProvider) newClientNova() (*gophercloud.ServiceClient, error) {
	handle, err := getHandle(p.config)
	if err != nil {
		return nil, err
	}
	client, err := getNovaClient(handle)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (p *OpenstackProvider) newClientGlance() (*gophercloud.ServiceClient, error) {
	handle, err := getHandle(p.config)
	if err != nil {
		return nil, err
	}
	client, err := getGlanceClient(handle)
	if err != nil {
		return nil, err
	}
	return client, nil
}
