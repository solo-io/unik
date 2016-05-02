package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
)

func (p *VirtualboxProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
		SupportedCompilers: []string{compilers.RUMP_GO_VIRTUALBOX},
	}
}
