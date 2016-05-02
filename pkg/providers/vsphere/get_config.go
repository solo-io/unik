package vsphere

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
)

func (p *VsphereProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
		SupportedCompilers: []string{compilers.RUMP_GO_VIRTUALBOX},
	}
}
