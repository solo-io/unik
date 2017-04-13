package vsphere

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *VsphereProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
