package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/client"
	"path/filepath"
	"fmt"
	"strings"
	"os/exec"
	"os"
	"github.com/emc-advanced-dev/unik/pkg/util/log"
)

var name, path, compiler, provider, runArgs string
var mountPoints []string
var force bool

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a unikernel image from source code files",
	Long: `Compiles source files into a runnable unikernel image.
	Image is compiled for a target infrastructure, specified with the flag -provider.


	Example usage:
		unik build -name myUnikernel -path ./myApp/src -compiler rump-xen -provider aws -mountpoint /foo -mountpoint /bar -args '-myParameter MYVALUE' -force

		# will create a unikernel named myUnikernel using the sources found in ./myApp/src,
		# compiled using rumprun for the xen hypervisor, targeting AWS infrastructure,
		# expecting a volume to be mounted at /foo at runtime,
		# expecting another volume to be mounted at /bar at runtime,
		# passing '-myParameter MYVALUE' as arguments to the application when it is run,
		# and deleting any previous existing instances and image for the name myUnikernel before compiling

	Another example (using only the required parameters):
		unik build -name anotherUnikernel -path ./anotherApp/src -compiler rump-vmware -provider vsphere
`,
	Run: func(cmd *cobra.Command, args []string) {
		if name == "" {
			logrus.Error("--name must be set")
			os.Exit(-1)
		}
		if path == "" {
			logrus.Error("--path must be set")
			os.Exit(-1)
		}
		if compiler == "" {
			logrus.Error("--compiler must be set")
			os.Exit(-1)
		}
		if provider == "" {
			logrus.Error("--provider must be set")
			os.Exit(-1)
		}
		logrus.WithFields(logrus.Fields{
			"name": name,
			"path": path,
			"compiler": compiler,
			"provider": provider ,
			"args": args,
			"mountPoints": mountPoints,
			"force": force,
		}).Infof("running unik build")
		readClientConfig()
		path = strings.TrimSuffix(path, "/")
		sourceTar := path + "/" + name + ".tar.gz"
		tarCommand := exec.Command("tar", "-zvcf", filepath.Base(sourceTar), "./")
		log.LogCommand(tarCommand, true)
		tarCommand.Dir = path
		logrus.Info("Tarring files: %s\n", tarCommand.Args)
		err := tarCommand.Run()
		//clean up artifacts even if we fail
		defer func() {
			err = os.RemoveAll(sourceTar)
			if err != nil {
				logrus.WithError(err).Error("could not clean up tarball at " + sourceTar)
				os.Exit(-1)
			}
			fmt.Printf("cleaned up tarball %s\n", sourceTar)
		}()

		fmt.Printf("App packaged as tarball: %s\n", sourceTar)
		if url == "" {
			url = clientConfig.DaemonUrl
		}
		image, err := client.UnikClient(url).Images().Build(name, sourceTar, compiler, provider, runArgs, mountPoints, force)
		if err != nil {
			logrus.WithError(err).Error("building image failed")
			os.Exit(-1)
		}
		printImages(image)
	},
}

func init() {
	RootCmd.AddCommand(buildCmd)
	buildCmd.Flags().StringVar(&name, "name", "", "<string,required> name to give the unikernel. must be unique")
	buildCmd.Flags().StringVar(&path, "path", "", "<string,required> path to root application sources folder")
	buildCmd.Flags().StringVar(&compiler, "compiler", "", "<string,required> name of the unikernel compiler to use")
	buildCmd.Flags().StringVar(&provider, "provider", "", "<string,required> name of the target infrastructure to compile for")
	buildCmd.Flags().StringVar(&runArgs, "args", "", "<string,optional> to be passed to the unikernel at runtime")
	buildCmd.Flags().StringSliceVar(&mountPoints, "mountpoint", []string{}, "<string,repeated> specify up to 8 mount points for volumes")
	buildCmd.Flags().BoolVar(&force, "force", false, "<bool, optional> force overwriting a previously existing")
}
