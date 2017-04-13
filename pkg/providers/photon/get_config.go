package photon

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *PhotonProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
