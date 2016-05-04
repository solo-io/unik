package virtualbox

import (
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/compilers"
)

func getStorageType(extraConfig types.ExtraConfig) string {
	var storageType string
	switch extraConfig[compilers.STORAGE_CONTROLLER_TYPE]{
	case compilers.SATA_Storage:
		storageType = compilers.SATA_Storage
	default:
		storageType = compilers.SCSI_Storage
	}
	return storageType
}