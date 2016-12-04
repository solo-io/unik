package cmd

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/client"
)

var describeCompilerCmd = &cobra.Command{
	Use:   "describe-compiler",
	Short: "Describe compiler usage",
	Long: `Describes compiler usage.
You must provide triple base-language-provider to specify what compiler to describe.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := func() error {
			// Determine host.
			if err := readClientConfig(); err != nil {
				return err
			}
			if host == "" {
				host = clientConfig.Host
			}
			logrus.WithField("host", host).Info("listing providers")

			// Validate input.
			if base == "" {
				return errors.New("--base must be set", nil)
			}
			if lang == "" {
				return errors.New("--language must be set", nil)
			}
			if provider == "" {
				return errors.New("--provider must be set", nil)
			}
			logrus.WithFields(logrus.Fields{
				"base":     base,
				"lang":     lang,
				"provider": provider,
			}).Info("describe compiler")

			// Ask daemon.
			description, err := client.UnikClient(host).DescribeCompiler(base, lang, provider)
			if err != nil {
				return err
			}

			// Print result to the console.
			fmt.Println(description)

			return nil
		}(); err != nil {
			logrus.Errorf("failed describing compiler: %v", err)
			os.Exit(-1)
		}
	},
}

func init() {
	RootCmd.AddCommand(describeCompilerCmd)
	describeCompilerCmd.Flags().StringVar(&base, "base", "", "<string,required> name of the unikernel base to use")
	describeCompilerCmd.Flags().StringVar(&lang, "language", "", "<string,required> language the unikernel source is written in")
	describeCompilerCmd.Flags().StringVar(&provider, "provider", "", "<string,required> name of the target infrastructure to compile for")
}
