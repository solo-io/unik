package common

import (
	"encoding/json"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/djannot/aws-sdk-go/aws"
	"github.com/djannot/aws-sdk-go/aws/session"
	"github.com/djannot/aws-sdk-go/service/s3"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"io"
	"os"
)

const (
	unik_hub_region = "us-east-1"
	unik_hub_bucket = "unik-hub"
	unik_image_info = "Unik-Image-Info"
)

func PullImage(config config.HubConfig, imageName string, writer io.Writer) (*types.Image, error) {
	//to trigger modified djannot/aws-sdk
	os.Setenv("S3_AUTH_PROXY_URL", config.URL)
	metadata, err := download(imageKey(config, imageName), config.Password, writer)
	if err != nil {
		return nil, errors.New("downloading image", err)
	}
	var image types.Image
	if err := json.Unmarshal([]byte(metadata), &image); err != nil {
		return nil, errors.New("unmarshalling metadata for image", err)
	}
	logrus.Infof("downloaded image %v", image)
	return &image, nil
}

func PushImage(config config.HubConfig, image *types.Image, imagePath string) error {
	//to trigger modified djannot/aws-sdk
	os.Setenv("S3_AUTH_PROXY_URL", config.URL)
	metadata, err := json.Marshal(image)
	if err != nil {
		return errors.New("converting image metadata to json", err)
	}
	//upload image
	reader, err := os.Open(imagePath)
	if err != nil {
		return errors.New("opening file", err)
	}
	defer reader.Close()
	fileInfo, err := reader.Stat()
	if err != nil {
		return errors.New("getting file info", err)
	}
	if err := upload(config, imageKey(config, image.Name), string(metadata), reader, fileInfo.Size()); err != nil {
		return errors.New("uploading image file", err)
	}
	logrus.Infof("Image %v pushed to %s", image, config.URL)
	return nil
}

func download(key, password string, writer io.Writer) (string, error) {
	params := &s3.GetObjectInput{
		Bucket:   aws.String(unik_hub_bucket),
		Key:      aws.String(key),
		Password: aws.String(password),
	}
	result, err := s3.New(session.New(&aws.Config{Region: aws.String(unik_hub_region)})).GetObject(params)
	if err != nil {
		return "", errors.New("failed to download from s3", err)
	}
	n, err := io.Copy(writer, result.Body)
	if err != nil {
		return "", errors.New("copying image bytes", err)
	}
	logrus.Infof("downloaded %v bytes", n)
	if result.Metadata[unik_image_info] == nil {
		return "", errors.New(fmt.Sprintf(unik_image_info+" was empty. full metadata: %+v", result.Metadata), nil)
	}
	return *result.Metadata[unik_image_info], nil
}

func upload(config config.HubConfig, key, metadata string, body io.ReadSeeker, length int64) error {
	params := &s3.PutObjectInput{
		Body:   body,
		Bucket: aws.String(unik_hub_bucket),
		Key:    aws.String(key),
		Metadata: map[string]*string{
			"unik-password": aws.String(config.Password),
			"unik-email":    aws.String(config.Username),
			"unik-access":   aws.String("public"),
			unik_image_info: aws.String(metadata),
		},
	}
	result, err := s3.New(session.New(&aws.Config{Region: aws.String(unik_hub_region)})).PutObject(params)
	if err != nil {
		return errors.New("uploading image to s3 backend", err)
	}
	logrus.Infof("uploaded %v bytes: %v", length, result)
	return nil
}

func imageKey(config config.HubConfig, imageName string) string {
	return "/" + config.Username + "/" + imageName + "/latest" //TODO: support image versioning
}
