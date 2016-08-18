package cmd

import (
	"fmt"

	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"github.com/spf13/cobra"
	"net/http"
	"strings"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: "Search available images in the targeted Unik Image Repository",
	Long: `
Usage:

unik search

  - or -

unik search --imageName <imageName>

Requires that you first authenticate to a unik image repository with 'unik login'`,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := getHubConfig()
		if err != nil {
			logrus.Fatal(err)
		}
		resp, body, err := lxhttpclient.Get(c.URL, "/images", nil)
		if err != nil {
			logrus.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			logrus.Fatal(fmt.Sprintf("failed with status %v: %s", resp.StatusCode, string(body)))
		}
		var images []*types.Image
		if err := json.Unmarshal(body, &images); err != nil {
			logrus.Fatal(err)
		}
		filteredImages := images[:0]
		if imageName != "" {
			for _, image := range images {
				if !strings.Contains(image.Name, imageName) {
					filteredImages = append(filteredImages, image)
				}
			}
		} else {
			filteredImages = images
		}
		printImages(filteredImages)
	},
}

func init() {
	RootCmd.AddCommand(searchCmd)
	pullCmd.Flags().StringVar(&imageName, "imageName", "", "<string,optional> search images by names containing this string")
}
