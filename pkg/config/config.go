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
	Xen        []Xen        `yaml:"xen"`
	Openstack  []Openstack  `yaml:"openstack"`
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

type Xen struct {
	Name       string `yaml:"name"`
	KernelPath string `yaml:"pv_kernel"`
	XenBridge  string `yaml:"xen_bridge"`
}

type Openstack struct {
	Name string `yaml:"name"`

	UserName   string `yaml:"username"`
	UserId     string `yaml:"userid"`
	Password   string `yaml:"password"`
	AuthUrl    string `yaml:"auth_url"`
	TenantId   string `yaml:"tenant_id"`
	TenantName string `yaml:"tenant_name"`
	DomainId   string `yaml:"domain_id"`
	DomainName string `yaml:"domain_name"`

	ProjectName string `yaml:"project_name"`
	RegionId    string `yaml:"region_id"`
	RegionName  string `yaml:"region_name"`
}

type VirtualboxAdapterType string

const (
	BridgedAdapter  = VirtualboxAdapterType("bridged")
	HostOnlyAdapter = VirtualboxAdapterType("host_only")
)

type ClientConfig struct {
	Host string `yaml:"host"`
}

type HubConfig struct {
	URL      string `yaml:"url",json:"url"`
	Username string `yaml:"user",json:"user"`
	Password string `yaml:"pass",json:"pass"`
}
