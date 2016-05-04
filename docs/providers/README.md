# Providers
**Providers** conform to the interface:
```
type Provider interface {
	GetConfig() ProviderConfig
	//Images
	Stage(params types.StageImageParams) (*types.Image, error)
	ListImages() ([]*types.Image, error)
	GetImage(nameOrIdPrefix string) (*types.Image, error)
	DeleteImage(id string, force bool) error
	//Instances
	RunInstance(params types.RunInstanceParams) (*types.Instance, error)
	ListInstances() ([]*types.Instance, error)
	GetInstance(nameOrIdPrefix string) (*types.Instance, error)
	DeleteInstance(id string, force bool) error
	StartInstance(id string) error
	StopInstance(id string) error
	GetInstanceLogs(id string) (string, error)
	//Volumes
	CreateVolume(params types.CreateVolumeParams) (*types.Volume, error)
	ListVolumes() ([]*types.Volume, error)
	GetVolume(nameOrIdPrefix string) (*types.Volume, error)
	DeleteVolume(id string, force bool) error
	AttachVolume(id, instanceId, mntPoint string) error
	DetachVolume(id string) error
}
```

**Providers** handle the long-term management of UniK's principle object types:
* Images
* Instances
* Volumes

Providers typically store some type of state, which may include a JSON representation of the existing state, as well as disk image files. UniK's default providers currently store their respective states in `~/.unik/`.

Providers perform API calls talk to the hypervisor / cloud provider / infrastructure where the images are hosted and instances are run.

To add an implemented provider to the Daemon, see the Unik API Server constructor function `func NewUnikDaemon(config config.DaemonConfig) (*UnikDaemon, error)` in [`daemon.go`](../pkg/daemon/daemon.go)

Your change should look something like this:
```
func NewUnikDaemon(config config.DaemonConfig) (*UnikDaemon, error) {
	_providers := make(providers.Providers)
    //...
	for _, awsConfig := range config.Providers.Aws {
		logrus.Infof("Bootstrapping provider %s with config %v", aws_provider, awsConfig)
		p := aws.NewAwsProvier(awsConfig)
		s, err := state.BasicStateFromFile(aws.AwsStateFile)
		if err != nil {
			logrus.WithError(err).Warnf("failed to read aws state file at %s, creating blank aws state", aws.AwsStateFile)
			s = state.NewBasicState(aws.AwsStateFile)
		}
		p = p.WithState(s)
		_providers[aws_provider] = p
		break
	}
  //...
  //Add your provider here, like so:
  /*
  myConfig := myprovider.GetMyConfig()
  myProvider, err := myprovider.NewProvider()
  if err != nil {
    //handle err
  }
  _providers["my_provider_name"] = myProvider
  */
  //...
  d := &UnikDaemon{
    server:    lxmartini.QuietMartini(),
    providers: _providers,
    compilers: _compilers,
  }
  return d, nil
}
```
