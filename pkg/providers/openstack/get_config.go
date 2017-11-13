package openstack

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *OpenstackProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
