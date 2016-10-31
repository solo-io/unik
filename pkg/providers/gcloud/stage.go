package gcloud

import (
	"github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/providers/common"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/emc-advanced-dev/unik/pkg/util"
	"github.com/pborman/uuid"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/storage/v1"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func (p *GcloudProvider) Stage(params types.StageImageParams) (_ *types.Image, err error) {
	images, err := p.ListImages()
	if err != nil {
		return nil, errors.New("retrieving image list for existing image", err)
	}

	for _, image := range images {
		if image.Name == params.Name {
			if !params.Force {
				return nil, errors.New("an image already exists with name '"+params.Name+"', try again with --force", nil)
			} else {
				logrus.WithField("image", image).Warnf("force: deleting previous image with name " + params.Name)
				err = p.DeleteImage(image.Id, true)
				if err != nil {
					return nil, errors.New("removing previously existing image", err)
				}
			}
		}
	}

	logrus.WithField("raw-image", params.RawImage).WithField("project id", p.config.ProjectID).Infof("creating google image from raw image")

	rawImageFile, err := os.Stat(params.RawImage.LocalImagePath)
	if err != nil {
		return nil, errors.New("statting raw image file", err)
	}

	imageSize := rawImageFile.Size()

	//need to convert image to raw & name it disk.raw before uploading
	if params.RawImage.StageSpec.ImageFormat != types.ImageFormat_RAW {
		rawImage, err := ioutil.TempFile("", "converted.raw.image.")
		if err != nil {
			return nil, errors.New("creating tmp file for raw image", err)
		}
		defer os.Remove(rawImage.Name())
		logrus.Debugf("need to convert %v to image format RAW", params.RawImage.StageSpec.ImageFormat)
		if err := common.ConvertRawImage(params.RawImage.StageSpec.ImageFormat, types.ImageFormat_RAW, params.RawImage.LocalImagePath, rawImage.Name()); err != nil {
			return nil, errors.New("converting qcow2 to vhd image", err)
		}
		os.Remove(params.RawImage.LocalImagePath)
		//point at the new image
		params.RawImage.LocalImagePath = rawImage.Name()
		params.RawImage.StageSpec.ImageFormat = types.ImageFormat_RAW
		imageSize, err = common.GetVirtualImageSize(params.RawImage.LocalImagePath, params.RawImage.StageSpec.ImageFormat)
		if err != nil {
			return nil, errors.New("getting virtual image size", err)
		}
	}
	destDir, err := ioutil.TempDir("", "gcloud.raw.image.dir.")
	if err != nil {
		return nil, errors.New("creating tmp dir for gcloud image upload", err)
	}
	if !params.NoCleanup {
		defer os.RemoveAll(destDir)
	}
	gcloudImageName := filepath.Join(destDir, "disk.raw")
	if err := os.Rename(params.RawImage.LocalImagePath, gcloudImageName); err != nil {
		return nil, errors.New("renaming image to disk.raw ", err)
	}
	//if we're on OSX, we need to tar with gtar
	tarBin := "tar"
	if runtime.GOOS == "darwin" {
		tarBin = "gtar"
		if _, err := exec.LookPath(tarBin); err != nil {
			return nil, errors.New("gtar was not found in your system path. GNU Tar is required for running google cloud provider; try installing with 'brew install gtar'", err)
		}
	}
	tarCmd := exec.Command(tarBin, "-Sczf", "raw-disk.tar.gz", "disk.raw")
	tarCmd.Dir = destDir
	util.LogCommand(tarCmd, true)
	if err := tarCmd.Run(); err != nil {
		return nil, errors.New("running tar command ", err)
	}
	objectName := "raw-disk.tar.gz"

	//create tmp bucket
	bucketName := "unik-tmp-bucket-" + uuid.New()

	if !params.NoCleanup {
		defer p.storage().Objects.Delete(bucketName, objectName)
		defer p.storage().Buckets.Delete(bucketName)
	}

	bucket, err := p.storage().Buckets.Insert(p.config.ProjectID, &storage.Bucket{Name: bucketName}).Do()
	if err != nil {
		return nil, errors.New("creating bucket "+bucketName, err)
	}
	logrus.Debug("created bucket ", bucket)

	imageTar := filepath.Join(destDir, objectName)
	file, err := os.Open(imageTar)
	if err != nil {
		return nil, errors.New("opening file "+imageTar, err)
	}
	obj, err := p.storage().Objects.Insert(bucket.Name, &storage.Object{Name: objectName}).Media(file).Do()
	if err != nil {
		return nil, errors.New("uploading file "+imageTar, err)
	}
	logrus.Debug("uploaded object ", obj)

	imageSpec := &compute.Image{
		Name:       params.Name,
		SourceDisk: "gs://" + bucketName + "/" + objectName,
	}

	gImage, err := p.compute().Images.Insert(p.config.ProjectID, imageSpec).Do()
	if err != nil {
		return nil, errors.New("creating gcloud image from storage", err)
	}

	logrus.Infof("created google image successfully: %+v", gImage)

	sizeMb := imageSize >> 20

	image := &types.Image{
		Id:             gImage.Id,
		Name:           params.Name,
		RunSpec:        params.RawImage.RunSpec,
		StageSpec:      params.RawImage.StageSpec,
		SizeMb:         sizeMb,
		Infrastructure: types.Infrastructure_AWS,
		Created:        time.Now(),
	}
	if err := p.state.ModifyImages(func(images map[string]*types.Image) error {
		images[image.Id] = image
		return nil
	}); err != nil {
		return nil, errors.New("modifying image map in state", err)
	}

	logrus.WithFields(logrus.Fields{"image": image}).Infof("image created succesfully")
	return image, nil
}
