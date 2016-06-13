package helpers

import (
	"os"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/pkg/errors"
	"path/filepath"
	"os/exec"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"io/ioutil"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"github.com/Sirupsen/logrus"
	"runtime"
	"fmt"
	"time"
)

type TempUnikHome struct {
	Dir string
}

func (t *TempUnikHome) SetupUnik() {
	if runtime.GOOS == "darwin" {
		tmpDir := filepath.Join(os.Getenv("HOME"), ".unik", "tmp")
		os.Setenv("TMPDIR", tmpDir)
		os.MkdirAll(tmpDir, 0755)
	}

	n, err := ioutil.TempDir("", "unikhome.")
	if err != nil {
		panic(err)
	}
	config.Internal.UnikHome = n

	t.Dir = n
}

func (t *TempUnikHome) TearDownUnik() {
	os.RemoveAll(t.Dir)
}

func requireEnvVar(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", errors.New(fmt.Sprintf("%s must be set", key), nil)
	}
	return val, nil
}

func NewAwsConfig() (_ config.Aws, err error) {
	region, err := requireEnvVar("AWS_REGION")
	if err != nil {
		return
	}
	zone, err := requireEnvVar("AWS_AVAILABILITY_ZONE")
	if err != nil {
		return
	}
	return config.Aws{
		Name: "TEST-AWS-CONFIG",
		Region: region,
		Zone: zone,
	}, nil
}

func NewVirtualboxConfig() (_ config.Virtualbox, err error) {
	adapterName, err := requireEnvVar("VBOX_ADAPTER_NAME")
	if err != nil {
		return
	}
	adapterType, err := requireEnvVar("VBOX_ADAPTER_TYPE")
	if err != nil {
		return
	}

	return config.Virtualbox{
		Name: "TEST-VBOX-CONFIG",
		AdapterName: adapterName,
		VirtualboxAdapterType: config.VirtualboxAdapterType(adapterType),
	}, nil
}

func NewVsphereConfig() (_ config.Vsphere, err error) {
	vsphereUser, err := requireEnvVar("VSPHERE_USERNAME")
	if err != nil {
		return
	}
	vspherePassword, err := requireEnvVar("VSPHERE_PASSWORD")
	if err != nil {
		return
	}
	vsphereUrl, err := requireEnvVar("VSPHERE_URL")
	if err != nil {
		return
	}
	vsphereDatastore, err := requireEnvVar("VSPHERE_DATASTORE")
	if err != nil {
		return
	}
	vsphereDatacenter, err := requireEnvVar("VSPHERE_DATACENTER")
	if err != nil {
		return
	}
	vsphereNetworkLabel, err := requireEnvVar("VSPHERE_NETWORK_LABEL")
	if err != nil {
		return
	}

	return config.Vsphere{
		Name: "TEST-VBOX-CONFIG",
		VsphereUser: vsphereUser,
		VspherePassword: vspherePassword,
		VsphereURL: vsphereUrl,
		Datastore: vsphereDatastore,
		Datacenter: vsphereDatacenter,
		NetworkLabel: vsphereNetworkLabel,
	}, nil
}

func ConfigWithAws(config config.DaemonConfig, aws config.Aws) (config.DaemonConfig) {
	config.Providers.Aws = append(config.Providers.Aws, aws)
	return config
}

func ConfigWithVirtualbox(config config.DaemonConfig, virtualbox config.Virtualbox) (config.DaemonConfig) {
	config.Providers.Virtualbox = append(config.Providers.Virtualbox, virtualbox)
	return config
}

func ConfigWithVsphere(config config.DaemonConfig, vsphere config.Vsphere) (config.DaemonConfig) {
	config.Providers.Vsphere = append(config.Providers.Vsphere, vsphere)
	return config
}

func NewTestConfig() (cfg config.DaemonConfig) {
	noConfig := true
	if os.Getenv("TEST_AWS") != "" && os.Getenv("TEST_AWS") != "0" {
		awsConfig, err := NewAwsConfig()
		if err != nil {
			logrus.Panic(err)
		}
		cfg = ConfigWithAws(cfg, awsConfig)
		noConfig = false
	}
	if os.Getenv("TEST_VIRTUALBOX") != "" && os.Getenv("TEST_VIRTUALBOX") != "0" {
		vboxConfig, err := NewVirtualboxConfig()
		if err != nil {
			logrus.Panic(err)
		}
		cfg = ConfigWithVirtualbox(cfg, vboxConfig)
		noConfig = false
	}
	if os.Getenv("TEST_VSPHERE") != "" && os.Getenv("TEST_VSPHERE") != "0" {
		vsphereConfig, err := NewVsphereConfig()
		if err != nil {
			logrus.Panic(err)
		}
		cfg = ConfigWithVsphere(cfg, vsphereConfig)
		noConfig = false
	}
	if noConfig {
		logrus.WithField("cfg", cfg).Panic("at least one config must be specified with TEST_<Provider>")
	}
	return
}

func MakeContainers(projectRoot string) error {
	cmd := exec.Command("make", "-C", projectRoot, "containers")
	util.LogCommand(cmd, true)
	return cmd.Run()
}

func RemoveContainers(projectRoot string) error {
	cmd := exec.Command("make", "-C", projectRoot, "remove-containers")
	util.LogCommand(cmd, false)
	return cmd.Run()
}

func TarExampleApp(appDir string) (*os.File, error) {
	projectRoot := GetProjectRoot()
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, errors.New("getting abs of "+projectRoot, err)
	}
	path := filepath.Join(absRoot, "docs", "examples", appDir)
	logrus.Debugf("tarring sources at %s", path)
	sourceTar, err := ioutil.TempFile("", "example.app.tar.gz.")
	if err != nil {
		return nil, errors.New("failed to create tmp tar file", err)
	}
	if err := unikos.Compress(path, sourceTar.Name()); err != nil {
		os.RemoveAll(path)
		return nil, errors.New("failed to tar sources", err)
	}
	return sourceTar, nil
}

func TarTestApp(appDir string) (*os.File, error) {
	projectRoot := GetProjectRoot()
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, errors.New("getting abs of "+projectRoot, err)
	}
	path := filepath.Join(absRoot, "test", "test_apps", appDir)
	logrus.Debugf("tarring sources at %s", path)
	sourceTar, err := ioutil.TempFile("", "test.app.tar.gz.")
	if err != nil {
		return nil, errors.New("failed to create tmp tar file", err)
	}
	if err := unikos.Compress(path, sourceTar.Name()); err != nil {
		return nil, errors.New("failed to tar sources", err)
	}
	return sourceTar, nil
}

func TarTestVolume() (*os.File, error) {
	projectRoot := GetProjectRoot()
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return nil, errors.New("getting abs of "+projectRoot, err)
	}
	path := filepath.Join(absRoot, "test", "test_apps", "test_volume")
	logrus.Debugf("tarring data at %s", path)
	dataTar, err := ioutil.TempFile("", "test.data.tar.gz.")
	if err != nil {
		return nil, errors.New("failed to create tmp tar file", err)
	}
	if err := unikos.Compress(path, dataTar.Name()); err != nil {
		return nil, errors.New("failed to tar data", err)
	}
	return dataTar, nil
}

func BuildExampleImage(daemonUrl, exampleName, compiler, provider string, mounts []string) (*types.Image, error) {
	force := true
	noCleanup := false
	testSourceTar, err := TarExampleApp(exampleName)
	if err != nil {
		return nil, errors.New("tarring example app", err)
	}
	defer os.RemoveAll(testSourceTar.Name())
	return client.UnikClient(daemonUrl).Images().Build(exampleName, testSourceTar.Name(), compiler, provider, "", mounts, force, noCleanup)
}

func BuildTestImage(daemonUrl, appDir, compiler, provider string, mounts []string) (*types.Image, error) {
	force := true
	noCleanup := false
	testSourceTar, err := TarTestApp(appDir)
	if err != nil {
		return nil, errors.New("tarring test app", err)
	}
	defer os.RemoveAll(testSourceTar.Name())
	return client.UnikClient(daemonUrl).Images().Build(appDir, testSourceTar.Name(), compiler, provider, "", mounts, force, noCleanup)
}

func RunExampleInstance(daemonUrl, instanceName, imageName string, mountPointsToVols map[string]string) (*types.Instance, error) {
	noCleanup := false
	env := map[string]string{"FOO": "BAR"}
	memoryMb := 128
	return client.UnikClient(daemonUrl).Instances().Run(instanceName, imageName, mountPointsToVols, env, memoryMb, noCleanup, false)
}

func CreateExampleVolume(daemonUrl, volumeName, provider string, size int) (*types.Volume, error) {
	return client.UnikClient(daemonUrl).Volumes().Create(volumeName, "", provider, size, false)
}

func CreateTestDataVolume(daemonUrl, volumeName, provider string) (*types.Volume, error) {
	dataTar, err := TarTestVolume()
	if err != nil {
		return nil, errors.New("tarring test data volume", err)
	}
	defer os.RemoveAll(dataTar.Name())
	return client.UnikClient(daemonUrl).Volumes().Create(volumeName, dataTar.Name(), provider, 0, false)
}

func GetProjectRoot() string {
	projectRoot := os.Getenv("PROJECT_ROOT")
	if projectRoot == "" {
		_, filename, _, ok := runtime.Caller(1)
		if !ok {
			logrus.Panic("could not get current file")
		}
		projectRoot = filepath.Join(filepath.Dir(filename), "..", "..")
	}
	logrus.Infof("using %s as project root", projectRoot)
	return projectRoot
}

func WaitForIp(daemonUrl, instanceId string, timeout time.Duration) (string, error) {
	errc := make(chan error)
	go func(){
		<-time.After(timeout)
		errc <- errors.New("getting instance ip timed out after "+timeout.String(), nil)
	}()

	resultc := make(chan string)
	go func(){
		logrus.Infof("retrieving ip for instance %s", instanceId)
		started := time.Now()
		end := started.Add(timeout)
		for {
			instance, err := client.UnikClient(daemonUrl).Instances().Get(instanceId)
			if err != nil {
				errc <- errors.New("getting instance from UniK daemon", err)
				return
			}
			if instance.IpAddress != "" {
				resultc <- instance.IpAddress
				return
			}
			logrus.Debugf("sleeping %v left...", end.Sub(time.Now()))
 			time.Sleep(time.Second)
		}
	}()
	select {
	case result := <- resultc:
		return result, nil
	case err := <- errc:
		return "", err
	}
}
