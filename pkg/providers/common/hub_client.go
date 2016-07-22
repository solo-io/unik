package common

import (
	"encoding/json"
	"github.com/Sirupsen/logrus"
	"github.com/djannot/aws-sdk-go/aws"
	"github.com/djannot/aws-sdk-go/aws/request"
	"github.com/djannot/aws-sdk-go/aws/session"
	"github.com/djannot/aws-sdk-go/private/signer/v4"
	"github.com/djannot/aws-sdk-go/service/s3"
	"github.com/djannot/aws-sdk-go/service/s3/s3manager"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"io"
	"io/ioutil"
	"os"
)

func PullImage(config config.HubConfig, imageName string, writeTo io.WriterAt) (*types.Image, error) {
	metadataFile, err := ioutil.TempFile("", imageName+"-metadata")
	if err != nil {
		return nil, errors.New("creating tmp metadata file", err)
	}
	defer os.RemoveAll(metadataFile.Name())
	if err := download(config, "unik-hub-"+config.Username, imageName+".metadata", metadataFile); err != nil {
		return nil, errors.New("downloading image metdata", err)
	}
	data, err := ioutil.ReadFile(metadataFile.Name())
	if err != nil {
		return nil, errors.New("reading metadata for "+imageName, err)
	}
	var image types.Image
	if err := json.Unmarshal(data, &image); err != nil {
		return nil, errors.New("unmarshalling metadata for image", err)
	}
	if err := download(config, "unik-hub-"+config.Username, imageName, writeTo); err != nil {
		return nil, errors.New("downloading image", err)
	}
	logrus.Infof("downloaded image %v", image)
	return &image, nil
}

func PushImage(config config.HubConfig, image *types.Image, imagePath string) error {
	//create bucket if it doesn't exist
	bucketName := "unik-hub-" + config.Username
	if err := createBucket(config.URL, bucketName); err != nil {
		logrus.Warnf("creating bucket: %v", err)
	}

	//upload metadata first
	data, err := json.Marshal(image)
	if err != nil {
		return errors.New("converting image metadata to json", err)
	}
	metadataFile, err := ioutil.TempFile("", "tmp-metadata")
	if err != nil {
		return errors.New("creating tmp metadata file", err)
	}
	if err := ioutil.WriteFile(metadataFile.Name(), data, 0644); err != nil {
		return errors.New("writing metadata to tmp file", err)
	}
	if err := upload(config, bucketName, image.Name+".metadata", metadataFile, int64(len(data))); err != nil {
		return errors.New("uploading metadata file", err)
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
	if err := upload(config, bucketName, image.Name, reader, fileInfo.Size()); err != nil {
		return errors.New("uploading image file", err)
	}

	logrus.Infof("Image %v pushed to %s", image, config.URL)
	return nil
}

func download(config config.HubConfig, bucketName, key string, writeTo io.WriterAt) error {
	params := &s3.GetObjectInput{
		Bucket:   aws.String(bucketName), // required
		Key:      aws.String(key),        // required
		Password: aws.String(config.Password),
	}
	downloader := s3manager.NewDownloader(session.New())
	downloader.RequestOption = func(req *request.Request) *request.Request {
		return sign(config, req)
	}
	n, err := downloader.Download(writeTo, params)
	if err != nil {
		return err
	}
	logrus.Infof("downloaded %v bytes", n)
	return nil
}

func upload(config config.HubConfig, bucketName, key string, body io.ReadSeeker, length int64) error {
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName), // required
		Key:           aws.String(key),        // required
		ACL:           aws.String("private"),
		Body:          body,
		ContentLength: aws.Int64(length),
		ContentType:   aws.String("application/octet-stream"),
	}
	req, _ := s3.New(session.New()).PutObjectRequest(params)
	req = sign(config, req)

	if err := req.Send(); err != nil {
		return errors.New("uploading file", err)
	}
	if req.Error != nil {
		return errors.New("put object failed", req.Error)
	}
	logrus.Infof("uploaded %v bytes", length)
	return nil
}

func createBucket(hubUrl, name string) error {
	if resp, _, err := lxhttpclient.Post(hubUrl, "/create_bucket/"+name, nil, nil); err != nil {
		return errors.New("performing post", err)
	} else if resp.StatusCode != 201 {
		return errors.New("expected 201", nil)
	}
	return nil
}

func sign(config config.HubConfig, req *request.Request) *request.Request {
	os.Setenv("S3_AUTH_PROXY_URL", config.URL)
	req.HTTPRequest.Header.Set("X-Amz-Meta-Unik-Email", config.Username)
	req.HTTPRequest.Header.Set("X-Amz-Meta-Unik-Password", config.Password)
	req.HTTPRequest.Header.Set("X-Amz-Meta-Unik-Access", "private")
	v4.Sign(req)
	return req
}
