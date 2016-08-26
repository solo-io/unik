package openstack

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	unikos "github.com/emc-advanced-dev/unik/pkg/os"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/flavors"
	"github.com/rackspace/gophercloud/openstack/imageservice/v2/images"
	"github.com/rackspace/gophercloud/pagination"
	"math"
	"os"
	"time"
)

func (p *OpenstackProvider) Stage(params types.StageImageParams) (_ *types.Image, err error) {
	imageList, err := p.ListImages()
	if err != nil {
		return nil, errors.New("failed to retrieve image list", err)
	}

	// Handle image name collision.
	for _, image := range imageList {
		if image.Name == params.Name {
			if !params.Force {
				return nil, errors.New(fmt.Sprintf("an image already exists with name '%s', try again with --force", params.Name), nil)
			} else {
				logrus.WithField("image", image).Warnf("force: deleting previous image with name '%s'", params.Name)
				err = p.DeleteImage(image.Id, true)
				if err != nil {
					return nil, errors.New("failed to remove existing image", err)
				}
			}
		}
	}

	clientGlance, err := p.newClientGlance()
	if err != nil {
		return nil, err
	}
	clientNova, err := p.newClientNova()
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"params": params,
	}).Info("creating boot image from raw image")

	rawImageFile, err := os.Stat(params.RawImage.LocalImagePath)
	if err != nil {
		return nil, errors.New("statting raw image file", err)
	}

	// TODO: Obtain image LOGICAL size, not actual (e.g. 10GB for OSv, not 8MB)
	imageSizeB := rawImageFile.Size()
	imageSizeMB := int(unikos.Bytes(imageSizeB).ToMegaBytes())

	// Pick flavor.
	flavor, err := pickFlavor(clientNova, imageSizeMB, 0)
	if err != nil {
		return nil, errors.New("failed to pick flavor", err)
	}

	logrus.WithFields(logrus.Fields{
		"imageSizeB":  imageSizeB,
		"imageSizeMB": imageSizeMB,
		"flavor":      flavor,
	}).Debug("picking flavor")

	// Push image to OpenStack.
	createdImage, err := pushImage(clientGlance, params.Name, params.RawImage.LocalImagePath, flavor)
	if err != nil {
		return nil, err
	}

	image := &types.Image{
		Id:             createdImage.ID,
		Name:           createdImage.Name,
		RunSpec:        params.RawImage.RunSpec,
		StageSpec:      params.RawImage.StageSpec,
		SizeMb:         int64(imageSizeMB),
		Infrastructure: types.Infrastructure_OPENSTACK,
		Created:        time.Now(),
	}

	// Update state.
	if err := p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[createdImage.ID] = image
		return nil
	}); err != nil {
		return nil, errors.New("failed to modify image map in state", err)
	}

	logrus.WithFields(logrus.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}

// pickFlavor picks flavor that best matches criteria (i.e. HDD size and RAM size).
// While diskMB is required, memoryMB is optional (set to -1 to ignore).
func pickFlavor(clientNova *gophercloud.ServiceClient, diskMB int, memoryMB int) (*flavors.Flavor, error) {
	if diskMB <= 0 {
		return nil, errors.New("Please specify disk size.", nil)
	}

	var flavs []flavors.Flavor = listFlavors(clientNova, int(math.Ceil(float64(diskMB)/1024)), memoryMB)

	// Find smallest flavor for given conditions.
	logrus.Infof("Find smallest flavor for conditions: diskMB >= %d AND memoryMB >= %d\n", diskMB, memoryMB)

	var bestFlavor flavors.Flavor
	var minDiffDisk int = -1
	var minDiffMem int = -1
	for _, f := range flavs {
		diffDisk := f.Disk*1024 - diskMB
		var diffMem int = 0 // 0 is best value
		if memoryMB > 0 {
			diffMem = f.RAM - memoryMB
		}

		if diffDisk >= 0 && // disk is big enough
			(minDiffDisk == -1 || minDiffDisk > diffDisk) && // disk is smaller than current best, but still big enough
			diffMem >= 0 && // memory is big enough
			(minDiffMem == -1 || minDiffMem >= diffMem) { // memory is smaller than current best, but still big enough
			bestFlavor, minDiffDisk, minDiffMem = f, diffDisk, diffMem
		}
	}
	if minDiffDisk == -1 {
		return nil, errors.New(fmt.Sprintf("No flavor fits required conditions: diskMB >= %d AND memoryMB >= %d\n", diskMB, memoryMB), nil)
	}
	return &bestFlavor, nil
}

// listFlavors returns list of all flavors.
func listFlavors(clientNova *gophercloud.ServiceClient, minDiskGB int, minMemoryMB int) []flavors.Flavor {
	var flavs []flavors.Flavor = make([]flavors.Flavor, 0)

	pagerFlavors := flavors.ListDetail(clientNova, flavors.ListOpts{
		MinDisk: minDiskGB,
		MinRAM:  minMemoryMB,
	})
	pagerFlavors.EachPage(func(page pagination.Page) (bool, error) {
		flavorList, _ := flavors.ExtractFlavors(page)

		for _, f := range flavorList {
			flavs = append(flavs, f)
		}

		return true, nil
	})
	return flavs
}

// pushImage first creates meta for image at OpenStack, then it sends binary data for it, the qcow2 image.
func pushImage(clientGlance *gophercloud.ServiceClient, imageName string, imageFilepath string, flavor *flavors.Flavor) (*images.Image, error) {
	// Create metadata (on OpenStack).
	createdImage, err := createImage(clientGlance, imageName, flavor)
	if err != nil {
		return nil, errors.New("failed to create OpenStack image meta", err)
	}

	// Send the image binary data to OpenStack
	if err = uploadImage(clientGlance, createdImage.ID, imageFilepath); err != nil {
		return nil, errors.New("failed to upload image binary to OpenStack", err)
	}

	return createdImage, nil
}

// createImage creates image metadata on OpenStack.
func createImage(clientGlance *gophercloud.ServiceClient, name string, flavor *flavors.Flavor) (*images.Image, error) {
	createdImage, err := images.Create(clientGlance, images.CreateOpts{
		Name:             name,
		DiskFormat:       "qcow2",
		ContainerFormat:  "bare",
		MinDiskGigabytes: flavor.Disk,
	}).Extract()

	if err != nil {
		return nil, errors.New("failed to create image", err)
	}
	return createdImage, nil
}

// uploadImage uploads image binary data to existing OpenStack image metadata.
func uploadImage(clientGlance *gophercloud.ServiceClient, imageId string, filepath string) error {
	logrus.WithFields(logrus.Fields{
		"filepath": filepath,
	}).Info("Uploading composed image to OpenStack")

	f, err := os.Open(filepath)
	if err != nil {
		return errors.New("failed to open image file", err)
	}
	defer f.Close()

	res := images.Upload(clientGlance, imageId, f)
	return res.Err
}
