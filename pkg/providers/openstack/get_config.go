package openstack

import (
	"github.com/cf-unik/unik/pkg/providers"
)

func (p *OpenstackProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
