package firecracker

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *FirecrackerProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
