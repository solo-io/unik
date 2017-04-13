package virtualbox

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *VirtualboxProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
