package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *VsphereProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
