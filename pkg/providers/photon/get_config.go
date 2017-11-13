package photon

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *PhotonProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
