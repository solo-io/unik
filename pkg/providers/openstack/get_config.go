package openstack

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *OpenstackProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
