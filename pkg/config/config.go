package config

type DaemonConfig struct {
	Providers Providers `yaml:"providers"`
	Version   string    `yaml:"version"`
}

type Providers struct {
	Aws        []Aws        `yaml:"aws"`
	Vsphere    []Vsphere    `yaml:"vsphere"`
	Virtualbox []Virtualbox `yaml:"virtualbox"`
	Qemu       []Qemu       `yaml:"qemu"`
	Photon     []Photon     `yaml:"photon"`
}

type Aws struct {
	Name   string `yaml:"name"`
	Region string `yaml:"region"`
	Zone   string `yaml:"zone"`
}

type Vsphere struct {
	Name            string `yaml:"name"`
	VsphereUser     string `yaml:"vsphere_user"`
	VspherePassword string `yaml:"vsphere_password"`
	VsphereURL      string `yaml:"vsphere_url"`
	Datastore       string `yaml:"datastore"`
	Datacenter      string `yaml:"datacenter"`
	NetworkLabel    string `yaml:"network"`
}

type Photon struct {
	Name      string `yaml:"name"`
	PhotonURL string `yaml:"photon_url"`
	ProjectId string `yaml:"project_id"`
}

type Virtualbox struct {
	Name                  string                `yaml:"name"`
	AdapterName           string                `yaml:"adapter_name"`
	VirtualboxAdapterType VirtualboxAdapterType `yaml:"adapter_type"`
}

type Qemu struct {
	Name         string `yaml:"name"`
	NoGraphic    bool   `yaml:"no_graphic"`
	DebuggerPort int    `yaml:"debugger_port"`
}

type VirtualboxAdapterType string

const (
	BridgedAdapter  = VirtualboxAdapterType("bridged")
	HostOnlyAdapter = VirtualboxAdapterType("host_only")
)

type ClientConfig struct {
	Host string `yaml:"host"`
}
