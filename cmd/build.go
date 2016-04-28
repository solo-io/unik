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
)

var name, path, compiler, provider, runArgs string
var mountPoints []string
var force bool

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a unikernel image from source code files",
	Long: `Compiles source files into a runnable unikernel image.
	Building an image uses the following flags:
		--name: <string,required> name to give the unikernel. must be unique
		--path: <string,required> path to root application sources folder
		--compiler: <string,required> name of the unikernel compiler to use
		 (run 'unik list-compilers' for a list of available compilers)
		--provider: <string,required> name of the target cloud provider, hypervisor,
		 or bare metal infrastructure this image is intended for
		 (run 'unik list-providers' for a list of available provider)
		--mountpoint: <string,repeated> specify up to 8 mount points for volumes
		 the unikernel should attempt to run at runtime
		--args: <string,optional> to be passed to the unikernel at runtime
		--force: <bool, optional> force overwriting a previously existing
		 unikernel with the same name warning: (this will delete all running
		 instances of the uniknel)

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
		tarCommand.Stdout = os.Stdout
		tarCommand.Stderr = os.Stderr
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
		image, err := client.UnikClient(clientConfig.DaemonUrl).Images().Build(name, sourceTar, compiler, provider, runArgs, mountPoints, force)
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
