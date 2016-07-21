package cmd

import (
	"bufio"
	"fmt"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"strings"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to a Unik Repository to pull & push images",
	Run: func(cmd *cobra.Command, args []string) {
		defaultUrl := "http://hub.projectunik.io"
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Unik Hub Repository URL [%v]: ", defaultUrl)
		url, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		url = strings.Trim(url, "\n")
		if len(url) < 1 {
			url = defaultUrl
		}
		fmt.Printf("Username: ")
		user, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		fmt.Printf("Password: ")
		pass, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		setHubConfig(url, strings.Trim(user, "\n"), strings.Trim(pass, "\n"))
		fmt.Printf("using url %v\n", url)
	},
}

func init() {
	RootCmd.AddCommand(loginCmd)
}

func setHubConfig(url, user, pass string) error {
	data, err := yaml.Marshal(config.HubConfig{URL: url, Username: user, Password: pass})
	if err != nil {
		return errors.New("failed to convert config to yaml string ", err)
	}
	if err := ioutil.WriteFile(clientConfigFile, data, 0644); err != nil {
		return errors.New("failed writing config to file "+clientConfigFile, err)
	}
	return nil
}
