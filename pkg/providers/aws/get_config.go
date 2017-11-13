package aws

import (
	"github.com/solo-io/unik/pkg/providers"
)

func (p *AwsProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: false,
	}
}
