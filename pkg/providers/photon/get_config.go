package photon

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *PhotonProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
