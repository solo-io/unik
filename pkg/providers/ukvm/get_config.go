package ukvm

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *UkvmProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
