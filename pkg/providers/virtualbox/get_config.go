package virtualbox

import "github.com/emc-advanced-dev/unik/pkg/providers"

func (p *VirtualboxProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
