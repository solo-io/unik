package ukvm

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *UkvmProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
