package qemu

import (
	"github.com/emc-advanced-dev/unik/pkg/providers"
)

func (p *QemuProvider) GetConfig() providers.ProviderConfig {
	return providers.ProviderConfig{
		UsePartitionTables: true,
	}
}
