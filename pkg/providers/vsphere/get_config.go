package vsphere

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *VsphereProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
