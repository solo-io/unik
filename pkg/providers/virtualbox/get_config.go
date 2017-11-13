package virtualbox

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *VirtualboxProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
