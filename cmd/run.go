package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"strings"
)

var imageName string
var volumes, envPairs []string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a unikernel instance from a compiled image",
	Long: `Deploys a running instance from a unik-compiled unikernel disk image.
	The instance will be deployed on the provider the image was compiled for.
	e.g. if the image was compiled for virtualbox, unik will attempt to deploy
	the image on the configured virtualbox environment.

	'unik run' requires a unik-managed volume (see 'unik volumes' and 'unik create volume')
	to be attached and mounted to each mount point specified at image compilation time.
	This means that if the image was compiled with two mount points, /data1 and /data2,
	'unik run' requires 2 available volumes to be attached to the instance at runtime, which
	must be specified with the flags --vol SOME_VOLUME_NAME:/data1 --vol ANOTHER_VOLUME_NAME:/data2
	If no mount points are required for the image, volumes cannot be attached.

	environment variables can be set at runtime through the use of the -env flag.

	Example usage:
		unik run -name newInstance -imageName myImage -vol myVol:/mount1 -vol yourVol:/mount2 -env foo=bar -env another=one

		# will create and run an instance of myImage on the provider environment myImage is compiled for
		# instance will be named newInstance
		# instance will attempt to mount unik-managed volume myVol to /mount1
		# instance will attempt to mount unik-managed volume yourVol to /mount2
		# instance will boot with env variable 'foo' set to 'bar'
		# instance will boot with env variable 'another' set to 'one'

		# note that run must take exactly one -vol argument for each mount point defined in the image specification
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			logrus.Error("--name must be set")
			os.Exit(-1)
		}
		if imageName == "" {
			logrus.Error("--imageName must be set")
			os.Exit(-1)
		}
		readClientConfig()
		if url == "" {
			url = clientConfig.DaemonUrl
		}

		mounts := make(map[string]string)
		for _, vol := range volumes {
			pair := strings.Split(vol, ":")
			if len(pair) != 2 {
				logrus.Errorf("invalid format for vol flag: %s", vol)
				os.Exit(-1)
			}
			volId := pair[0]
			mnt := pair[1]
			mounts[volId] = mnt
		}

		env := make(map[string]string)
		for _, e := range envPairs {
			pair := strings.Split(e, "=")
			if len(pair) != 2 {
				logrus.Errorf("invalid format for env flag: %s", e)
				os.Exit(-1)
			}
			key := pair[0]
			val := pair[1]
			env[key] = val
		}

		logrus.WithFields(logrus.Fields{
			"name": name,
			"imageName": imageName,
			"env": env,
			"mounts": mounts,
		}).Infof("running unik run")
		instance, err := client.UnikClient(url).Instances().Run(name, imageName, mounts, env)
		if err != nil {
			logrus.WithError(err).Error("building image failed")
			os.Exit(-1)
		}
		printInstances(instance)
	},
}

func init() {
	RootCmd.AddCommand(runCmd)
	buildCmd.Flags().StringVar(&name, "name", "", "<string,required> name to give the instance. must be unique")
	buildCmd.Flags().StringVar(&imageName, "imageName", "", "<string,required> image to use")
	buildCmd.Flags().StringSliceVar(&envPairs, "env", "", "<string,repeated> set any number of environment variables for the instance. must be in the format KEY=VALUE")
	buildCmd.Flags().StringSliceVar(&volumes, "vol", "", `<string,repeated> each --vol flag specifies one volume id and the corresponding mount point to attach
	to the instance at boot time. volumes must be attached to the instance for each mount point expected by the image.
	run 'unik image <image_name>' to see the mount points required for the image.
	specified in the format 'volume_id:mount_point'`)
}
