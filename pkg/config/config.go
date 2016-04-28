package config

type DaemonConfig struct {
	Providers Providers `yaml:"providers"`
	Version   string    `yaml:"version"`
}
type Providers struct {
	Aws        []Aws        `yaml:"aws"`
	Vsphere    []Vsphere    `yaml:"vsphere"`
	Virtualbox []Virtualbox `yaml:"virtualbox"`
}

type Aws struct {
	Name              string `yaml:"name"`
	AwsAccessKeyID    string `yaml:"aws_access_key_id"`
	AwsSecretAcessKey string `yaml:"aws_secret_acess_key"`
	Region            string `yaml:"region"`
	Zone              string `yaml:"zone"`
}

type Vsphere struct {
	Name            string `yaml:"name"`
	VsphereUser     string `yaml:"vsphere_user"`
	VspherePassword string `yaml:"vsphere_password"`
	VsphereURL      string `yaml:"vsphere_url"`
	Datastore       string `yaml:"datastore"`
	DefaultInstanceMemory int `yaml:"default_instance_memory"`
}

type Virtualbox struct {
	Name                  string                `yaml:"name"`
	AdapterName           string                `yaml:"adapter_name"`
	VirtualboxAdapterType VirtualboxAdapterType `yaml:"adapter_type"`
}

type VirtualboxAdapterType string

const (
	BridgedAdapter  = VirtualboxAdapterType("bridged")
	HostOnlyAdapter = VirtualboxAdapterType("host_only")
)

type ClientConfig struct {
	DaemonUrl string `yaml:"url"`
}
