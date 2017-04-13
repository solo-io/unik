package xen

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *XenProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
