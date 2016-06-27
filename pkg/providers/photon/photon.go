package photon

import (
	"net/url"

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
