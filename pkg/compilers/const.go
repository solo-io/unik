package compilers

const (
	//available compilers
	RUMP_GO_AWS = "rump-go-aws"
	RUMP_GO_VMWARE = "rump-go-vmware"
	RUMP_GO_VIRTUALBOX = "rump-go-virtualbox"

	OSV_JAVA_AWS = "osv-java-aws"
	OSV_JAVA_VMAWRE = "osv-java-vmware"
	OSV_JAVA_VIRTUALBOX = "osv-java-virtualbox"
)

//config types
const (
	//Storage Config
	STORAGE_CONTROLLER_TYPE = "STORAGE_CONTROLLER_TYPE"

	SCSI_Storage = "SCSI_Storage"
	SATA_Storage = "SATA_Storage"

	//qemu-img convert config
	IMAGE_TYPE = "IMAGE_TYPE"

	RAW = "raw"
	VMDK = "vmdk"
	QCOW2 = "qcow2"
)