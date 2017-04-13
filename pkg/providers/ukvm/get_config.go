package ukvm

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *UkvmProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
