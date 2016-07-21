package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/djannot/aws-sdk-go/aws"
	"github.com/djannot/aws-sdk-go/aws/request"
	"github.com/djannot/aws-sdk-go/aws/session"
	"github.com/djannot/aws-sdk-go/private/signer/v4"
	"github.com/djannot/aws-sdk-go/service/s3"
	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/config"
	"github.com/layer-x/layerx-commons/lxhttpclient"
	"os"
	"github.com/djannot/aws-sdk-go/service/s3/s3manager"
)

func PullImage(config config.HubConfig, imageName, imagePath string) error {
	bucketName := "unik-hub-" + config.Username

	file, err := os.Open(imagePath)
	if err != nil {
		return errors.New("Failed to open file", err)
	}
	defer file.Close()

	// download
	params := &s3.GetObjectInput{
		Bucket:   aws.String(bucketName), // required
		Key:      aws.String(imageName),  // required
		Password: aws.String(config.Password),
	}
	downloader := s3manager.NewDownloader(session.New())
	n, err := downloader.Download(file, )
	req, out := s3.New(session.New()).PutObjectRequest(params)
	req = sign(config, req)

	if err := req.Send(); err != nil {
		return errors.New("downloading file", err)
	}
	if req.Error != nil {
		return errors.New("get object failed", req.Error)
	}

	logrus.Infof("Image saved to %s", imagePath)
	return nil
}

func PushImage(config config.HubConfig, imageName, imagePath string) error {
	//create bucket if it doesn't exist
	bucketName := "unik-hub-" + config.Username
	if err := createBucket(config.URL, bucketName); err != nil {
		logrus.Warnf("creating bucket: %v", err)
	}

	//read image file in
	reader, err := os.Open(imagePath)
	if err != nil {
		return errors.New("opening file", err)
	}
	defer reader.Close()

	fileInfo, err := reader.Stat()
	if err != nil {
		return errors.New("getting file info", err)
	}

	// upload
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName), // required
		Key:           aws.String(imageName),  // required
		ACL:           aws.String("private"),
		Body:          reader,
		ContentLength: aws.Int64(fileInfo.Size()),
		ContentType:   aws.String("application/octet-stream"),
	}
	req, _ := s3.New(session.New()).PutObjectRequest(params)
	req := sign(config, req)

	if err := req.Send(); err != nil {
		return errors.New("uploading file", err)
	}
	if req.Error != nil {
		return errors.New("put object failed", req.Error)
	}
	logrus.Infof("Image saved to %s", imageName)
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
