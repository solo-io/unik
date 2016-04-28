package cmd

import (
	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"os"
	"fmt"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"bytes"
)

var clientConfigFile, url string
var port int

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "unik",
	Short: "The unikernel compilation, deployment, and management tool",
	Long: `Unik is a tool for compiling application source code
	into bootable disk images. Unik also runs and manages unikernel
	instances across infrastructures.

	Set client configuration file with --client-config=<path>`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(-1)
	}
}

func init() {
	RootCmd.PersistentFlags().StringVar(&clientConfigFile, "client-config", os.Getenv("HOME")+"/.unik/client-config.yaml", "client config file (default is $HOME/.unik/client-config.yaml)")
	RootCmd.PersistentFlags().StringVar(&url, "url", "", "<string, optional>: url/ip address of the host running the unik daemon")
	targetCmd.Flags().IntVar(&port, "port", 3000, "<int, optional>: port the daemon is running on (default: 3000)")
}

var clientConfig config.ClientConfig
func readClientConfig() {
	data, err := ioutil.ReadFile(clientConfigFile)
	if err != nil {
		logrus.WithError(err).Errorf("failed to read client configuration file at "+ clientConfigFile +`\n
		Try setting your config with 'unik target DAEMON_URL'`)
		os.Exit(-1)
	}
	data = bytes.Replace(data, []byte("\n"), []byte{}, -1)
	if err := yaml.Unmarshal(data, &clientConfig); err != nil {
		logrus.WithError(err).Errorf("failed to parse client configuration yaml at "+ clientConfigFile +`\n
		Please ensure config file contains valid yaml.'`)
		os.Exit(-1)
	}
}

func printImages(images ... *types.Image) {
	fmt.Printf("%-20s %-20s %-20s %-20s %-6s %-20s\n", "NAME", "ID", "INFRASTRUCTURE", "CREATED", "SIZE", "MOUNTPOINTS")
	for _, image := range images {
		printImage(image)
	}
}

func printImage(image *types.Image) {
	if len(image.DeviceMappings) == 0 {
		fmt.Printf("%-20.20s %-20.20s %-14.14s %-20.20s %-6.6v \n", image.Name, image.Id, image.Infrastructure, image.Created.String(), image.SizeMb)
	} else if len(image.DeviceMappings) > 0 {
		fmt.Printf("%-20.20s %-20.20s %-14.14s %-20.20s %-6.6v %-20.20s\n", image.Name, image.Id, image.Infrastructure, image.Created.String(), image.SizeMb, image.DeviceMappings[0].MountPoint)
		if len(image.DeviceMappings) > 1 {
			for i := 1; i < len(image.DeviceMappings); i++ {
				fmt.Printf("%94.90s\n", image.DeviceMappings[i].MountPoint)
			}
		}
	}
}

func printInstances(instance ... *types.Instance) {
	fmt.Printf("%-15s %-20s %-14s %-20s %-20s %-14s %-12s\n",
		"NAME", "ID", "INFRASTRUCTURE", "CREATED", "IMAGE", "IPADDRESS", "STATE")
	for _, image := range instance {
		printInstance(image)
	}
}

func printInstance(instance *types.Instance) {
	fmt.Printf("%-15.15s %-20.20s %-14.14s %-20.20s %-20.20v %-14.14s %-12.12s\n",
		instance.Name, instance.Id, instance.Infrastructure, instance.Created.String(), instance.ImageId, instance.IpAddress, instance.State)
}