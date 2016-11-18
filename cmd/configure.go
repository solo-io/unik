package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"path/filepath"
	"github.com/emc-advanced-dev/pkg/errors"
)

var configureCmd = &cobra.Command{
	Use:   "configure [--provider PROVIDER-NAME]",
	Short: "A generate configuration file for daemon ('daemon.yaml')",
	Long:  `An interactive command to help walk you through the process of creating or changing a configuration file for unik.
Can be used to configure an individual provider, or any number of providers.

Usage:
unik configure # will iterate through all possible providers and ask if user wants to configure
-or-
unik configure --provider PROVIDER
where provider is one of the following:
aws
gcloud
openstack
photon
qemu
ukvm
virtualbox
vsphere
xen

	`,
	Run: func(cmd *cobra.Command, args []string) {
		readDaemonConfig()
		reader := bufio.NewReader(os.Stdin)
		var configFunc func() error
		switch provider {
		case "":
			configFunc = func() error {
				doAwsConfig(reader)
				doGcloudConfig(reader)
				doOpenstackConfig(reader)
				return nil
			}
		}
		if err := configFunc(); err != nil {
			logrus.Fatal(err)
		}
		if err := writeDaemonConfig(); err != nil {
			logrus.Fatal(err)
		}

	},
}

func init() {
	RootCmd.AddCommand(configureCmd)
	buildCmd.Flags().StringVar(&provider, "provider", "", "<string,optional> provider to configure. if not given, unik will iterate through each possible provider to configure")
}

func writeDaemonConfig() error {
	data, err := yaml.Marshal(daemonConfig)
	if err != nil {
		return errors.New("failed to convert config to yaml string ", err)
	}
	os.MkdirAll(filepath.Dir(daemonConfigFile), 0755)
	if err := ioutil.WriteFile(daemonConfigFile, data, 0644); err != nil {
		return errors.New("failed writing config to file "+daemonConfigFile, err)
	}
	return nil
}


func doAwsConfig(reader *bufio.Reader) error {
	fmt.Print("Do you wish to configure unik for use with AWS? [y/N]: ")
	y, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if y == "y" {
		if len(daemonConfig.Providers.Aws) < 1 {
			daemonConfig.Providers.Aws = append(daemonConfig.Providers.Aws, config.Aws{})
		}
		if daemonConfig.Providers.Aws[0].Name == "" {
			daemonConfig.Providers.Aws[0].Name = "aws-configuration"
		}
		fmt.Printf("AWS region where to deploy unikernels [%s]: ", daemonConfig.Providers.Aws[0].Region)
		region, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if region != "" {
			daemonConfig.Providers.Aws[0].Region = region
		}
		fmt.Printf("AWS availability zone where to deploy unikernels (must be within region) [%s]: ", daemonConfig.Providers.Aws[0].Zone)
		zone, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if zone != "" {
			daemonConfig.Providers.Aws[0].Zone = zone
		}
	}
	return nil
}

func doGcloudConfig(reader *bufio.Reader) error {
	fmt.Print("Do you wish to configure unik for use with Google Compute Engine? [y/N]: ")
	y, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if y == "y" {
		if len(daemonConfig.Providers.Gcloud) < 1 {
			daemonConfig.Providers.Aws = append(daemonConfig.Providers.Gcloud, config.Gcloud{})
		}
		if daemonConfig.Providers.Gcloud[0].Name == "" {
			daemonConfig.Providers.Gcloud[0].Name = "gcloud-configuration"
		}
		fmt.Printf("GCloud project id under which to deploy unikernels [%s]: ", daemonConfig.Providers.Gcloud[0].ProjectID)
		projectId, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if projectId != "" {
			daemonConfig.Providers.Gcloud[0].ProjectID = projectId
		}
		fmt.Printf("Gcloud availability zone where to deploy unikernels (must be within region) [%s]: ", daemonConfig.Providers.Gcloud[0].Zone)
		zone, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if zone != "" {
			daemonConfig.Providers.Gcloud[0].Zone = zone
		}
	}
	return nil
}

func doOpenstackConfig(reader *bufio.Reader) error {
	fmt.Print("Do you wish to configure unik for use with Openstack? [y/N]: ")
	y, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	if y == "y" {
		if len(daemonConfig.Providers.Openstack) < 1 {
			daemonConfig.Providers.Aws = append(daemonConfig.Providers.Openstack, config.Openstack{})
		}
		if daemonConfig.Providers.Openstack[0].Name == "" {
			daemonConfig.Providers.Openstack[0].Name = "Openstack-configuration"
		}
		fmt.Printf("Openstack username for authentication [%s]: ", daemonConfig.Providers.Openstack[0].UserName)
		username, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if username != "" {
			daemonConfig.Providers.Openstack[0].UserName = username
		}
		fmt.Printf("Openstack password for authentication [%s]: ", daemonConfig.Providers.Openstack[0].Password)
		password, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if password != "" {
			daemonConfig.Providers.Openstack[0].Password = password
		}
		fmt.Printf("Openstack authentication url [%s]: ", daemonConfig.Providers.Openstack[0].AuthUrl)
		authUrl, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if authUrl != "" {
			daemonConfig.Providers.Openstack[0].AuthUrl = authUrl
		}
		fmt.Printf("Openstack tenant id [%s]: ", daemonConfig.Providers.Openstack[0].TenantId)
		tenantId, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if tenantId != "" {
			daemonConfig.Providers.Openstack[0].TenantId = tenantId
		}
		fmt.Printf("Openstack project name [%s]: ", daemonConfig.Providers.Openstack[0].ProjectName)
		projectName, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if projectName != "" {
			daemonConfig.Providers.Openstack[0].ProjectName = projectName
		}
		fmt.Printf("Openstack region id [%s]: ", daemonConfig.Providers.Openstack[0].RegionId)
		regionId, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if regionId != "" {
			daemonConfig.Providers.Openstack[0].RegionId = regionId
		}
		fmt.Printf("Openstack network uuid [%s]: ", daemonConfig.Providers.Openstack[0].NetworkUUID)
		networkUUID, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		if networkUUID != "" {
			daemonConfig.Providers.Openstack[0].NetworkUUID = networkUUID
		}
	}
	return nil
}
