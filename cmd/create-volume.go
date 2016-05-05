package cmd

import (
	"io/ioutil"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/client"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	unikutil "github.com/emc-advanced-dev/unik/pkg/util"
)

var data string
var size int

var cvCmd = &cobra.Command{
	Use:   "create-volume",
	Short: "Create a unik-managed data volume",
	Long: `Create a data volume which can be attached to and detached from
unik-managed instances.

Volumes can be created from a directory, which will copy the contents
of the directory onto the voume. Empty volume can also be created.

Volumes will persist after instances are deleted, allowing application data
to be persisted beyond the lifecycle of individual instances.

If specifying a data folder (with --data), specifying a size for the volume is
not necessary. UniK will automatically size the volume to fit the data provided.
A larger volume can be requested with the --size flag.

If no data directory is provided, --size is a required parameter to specify the
desired size for the empty volume to be createad.

Volumes are created for a specific provider, specified with the --provider flag.
Volumes can only be attached to instances of the same provider type.
To see a list of available providers, run 'unik providers'

Volume names must be unique. If a volume exists with the same name, you will be
required to remove the volume with 'unik delete-volume' before the new volume
can be created.

--size parameter uses MB

Example usage:
	unik create-volume --name myVolume --data ./myApp/data --provider aws

	# will create an EBS-backed AWS volume named myVolume using the data found in ./myApp/src,
	# the size will be either 1GB (the default minimum size on AWS) or greater, if the size of the
	volume is greater


Another example (empty volume):
	unik create-volume -name anotherVolume --size 500 -provider vsphere

	# will create a 500mb sparse vmdk file and upload it to the vsphere datastore,
	where it can be attached to a vsphere instance
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if name == "" {
				return errors.New("--name must be set", nil)
			}
			if data == "" && size == 0 {
				return errors.New("either --data or --size must be set", nil)
			}
			if provider == "" {
				return errors.New("--provider must be set", nil)
			}
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithFields(logrus.Fields{
				"name":     name,
				"data":     data,
				"size":     size,
				"provider": provider,
				"host":     host,
			}).Infof("creating volume")
			if data != "" {
				dataTar, err := ioutil.TempFile(unikutil.UnikTmpDir(), "")
				if err != nil {
					logrus.WithError(err).Error("failed to create tmp tar file")
				}
				if false {
					defer os.Remove(dataTar.Name())
				}
				if err := unikos.Compress(path, dataTar.Name()); err != nil {
					return errors.New("failed to tar data", err)
				}
				data = dataTar.Name()
				logrus.Infof("Data packaged as tarball: %s\n", dataTar.Name())
			}
			volume, err := client.UnikClient(host).Volumes().Create(name, data, provider, size, noCleanup)
			if err != nil {
				return errors.New("creatinv volume image failed", err)
			}
			printVolumes(volume)
			return nil
		}(); err != nil {
			logrus.Errorf("create-volume failed: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(cvCmd)
	cvCmd.Flags().StringVar(&name, "name", "", "<string,required> name to give the unikernel. must be unique")
	cvCmd.Flags().StringVar(&data, "data", "", "<string,special> path to data folder. optional if --size is provided")
	cvCmd.Flags().IntVar(&size, "compiler", 0, "<int,special> size to create volume in MB. optional if --data is provided")
	cvCmd.Flags().StringVar(&provider, "provider", "", "<string,required> name of the target infrastructure to compile for")
	cvCmd.Flags().BoolVar(&noCleanup, "no-cleanup", false, "<bool, optional> for debugging; do not clean up artifacts for volumes that fail to build")
}
