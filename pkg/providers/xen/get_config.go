package xen

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *XenProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
