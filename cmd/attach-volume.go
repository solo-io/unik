package cmd

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/solo-io/unik/pkg/client"
)

var mountPoint string

var attachCmd = &cobra.Command{
	Use:     "attach-volume",
	Aliases: []string{"attach"},
	Short:   "Attach a volume to a stopped instance",
	Long: `Attaches a volume to a stopped instance at a specified mount point.
You specify the volume by name or id.

The volume must be attached to an available mount point on the instance.
Mount points are image-specific, and are determined when the image is compiled.

For a list of mount points on the image for this instance, run unik images, or
unik describe image

If the specified mount point is occupied by another volume, the command will result
in an error
`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			if err := readClientConfig(); err != nil {
				return err
			}
			if volumeName == "" {
				return errors.New("must specify --volume", nil)
			}
			if instanceName == "" {
				return errors.New("must specify --instanceName", nil)
			}
			if mountPoint == "" {
				return errors.New("must specify --mountPoint", nil)
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithFields(logrus.Fields{"host": host, "instanceName": instanceName, "volume": volumeName, "mountPoint": mountPoint}).Info("attaching volume")
			if err := client.UnikClient(host).Volumes().Attach(volumeName, instanceName, mountPoint); err != nil {
				return err
			}
			return nil
		}(); err != nil {
			logrus.Errorf("failed deleting volume: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(attachCmd)
	attachCmd.Flags().StringVar(&volumeName, "volume", "", "<string,required> name or id of volume to attach. unik accepts a prefix of the name or id")
	attachCmd.Flags().StringVar(&instanceName, "instance", "", "<string,required> name or id of instance to attach to. unik accepts a prefix of the name or id")
	attachCmd.Flags().StringVar(&mountPoint, "mountPoint", "", "<string,required> mount path for volume. this should reflect the mappings specified on the image. run 'unik describe-image' to see expected mount points for the image")
	attachCmd.Flags().BoolVar(&force, "force", false, "<bool, optional> force deleting volume in the case that it is running")
}
