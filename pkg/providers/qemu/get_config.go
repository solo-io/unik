package qemu

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
)

func (p *QemuProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
		SupportedCompilers: []string{
			compilers.RUMP_GO_QEMU,
			compilers.RUMP_NODEJS_QEMU,
			compilers.RUMP_PYTHON_QEMU,
		},
	}
}
