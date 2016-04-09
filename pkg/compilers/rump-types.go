package compilers

type Blk struct {
	Source     string `json:"source"`
	Path       string `json:"path"`
	FSType     string `json:"fstype"`
	MountPoint string `json:"mountpoint,omitempty"`
	DiskFile   string `json:"diskfile,omitempty"`
}

type Method string

const (
	Static Method = "static"
	DHCP   Method = "dhcp"
)

type Net struct {
	If     string `json:"if,omitempty"`
	Type   string `json:"type,omitempty"`
	Method Method `json:"method,omitempty"`
	Addr   string `json:"addr,omitempty"`
	Mask   string `json:"mask,omitempty"`
	Cloner string `json:"cloner,omitempty"`
}

type RumpConfig struct {
	Cmdline string `json:"cmdline"`
	Net     *Net   `json:"net,omitempty"`
	Blk     []Blk  `json:"blk,omitempty"`
}
