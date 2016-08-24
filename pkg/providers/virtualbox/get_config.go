package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/compilers"
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *VirtualboxProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
		SupportedCompilers: []string{
			compilers.RUMP_GO_VIRTUALBOX,
			compilers.RUMP_NODEJS_VIRTUALBOX,
			compilers.RUMP_PYTHON_VIRTUALBOX,
			compilers.RUMP_JAVA_VIRTUALBOX,
			compilers.OSV_JAVA_VIRTUALBOX,
			compilers.INCLUDEOS_CPP_VIRTUALBOX,
		},
	}
}
