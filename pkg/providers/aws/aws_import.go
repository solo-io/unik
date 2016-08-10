package aws

import (
	"encoding/xml"
	"os"
	"time"

	"bytes"
	"io"

	"fmt"

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"

	"math/rand"

	"github.com/emc-advanced-dev/pkg/errors"
	"github.com/emc-advanced-dev/unik/pkg/types"
	"strings"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func createDataVolumeFromRawImage(s3svc *s3.S3, ec2svc *ec2.EC2, imgFile string, imageSize int64, imageFormat types.ImageFormat, az string) (string, error) {
	fileInfo, err := os.Stat(imgFile)
	if err != nil {
		return "", err
	}

	// upload the image file to aws
	bucket := fmt.Sprintf("unik-tmp-%d", rand.Int63())

	if err := createBucket(s3svc, bucket); err != nil {
		return "", err
	}
	defer deleteBucket(s3svc, bucket)

	pathInBucket := "disk.img"

	log.Debug("Uploading image to aws")

	if err := uploadFileToAws(s3svc, imgFile, fileInfo.Size(), bucket, pathInBucket); err != nil {
		return "", err
	}

	log.Debug("Creating self sign urls")

	// create signed urls for the file (get, head, delete)
	// s.s3svc.

	getReq, _ := s3svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(pathInBucket),
	})
	getUrlStr, err := getReq.Presign(24 * time.Hour)
	if err != nil {
		return "", err
	}

	headReq, _ := s3svc.HeadObjectRequest(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(pathInBucket),
	})

	headUrlStr, err := headReq.Presign(24 * time.Hour)
	if err != nil {
		return "", err
	}

	deleteReq, _ := s3svc.DeleteObjectRequest(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(pathInBucket),
	})

	deleteUrlStr, err := deleteReq.Presign(24 * time.Hour)
	if err != nil {
		return "", err
	}

	log.Debug("Creating manifest")

	// create manifest
	manifestName := "upload-manifest.xml"

	deleteManiReq, _ := s3svc.DeleteObjectRequest(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(manifestName),
	})

	deleteManiUrlStr, err := deleteManiReq.Presign(24 * time.Hour)
	if err != nil {
		return "", err
	}

	m := manifest{
		Version:         "2010-11-15",
		FileFormat:      strings.ToUpper(string(imageFormat)),
		Importer:        importer{"unik", "1", "2016-04-01"},
		SelfDestructUrl: deleteManiUrlStr,
		ImportSpec: importSpec{
			Size:       fileInfo.Size(),
			VolumeSize: toGigs(imageSize),
			Parts: parts{
				Count: 1,
				Parts: []part{
					part{
						Index: 0,
						ByteRange: byteRange{
							Start: 0,
							End:   fileInfo.Size(),
						},
						Key:       pathInBucket,
						HeadUrl:   headUrlStr,
						GetUrl:    getUrlStr,
						DeleteUrl: deleteUrlStr,
					},
				},
			},
		},
	}
	// write manifest
	buf := new(bytes.Buffer)
	enc := xml.NewEncoder(buf)
	if err := enc.Encode(m); err != nil {
		return "", err
	}
	log.Debug("Uploading manifest")

	// upload manifest
	manifestBytes := buf.Bytes()
	err = uploadToAws(s3svc, bytes.NewReader(manifestBytes), int64(len(manifestBytes)), bucket, manifestName)
	if err != nil {
		return "", err
	}

	getManiReq, _ := s3svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(manifestName),
	})
	getManiUrlStr, err := getManiReq.Presign(24 * time.Hour)
	if err != nil {
		return "", err
	}

	log.Debug("Importing volume")

	// finally import the image
	volparams := &ec2.ImportVolumeInput{
		AvailabilityZone: aws.String(az), // Required
		Image: &ec2.DiskImageDetail{ // Required
			Bytes:             aws.Int64(toGigs(imageSize)),                     // Required
			Format:            aws.String(strings.ToUpper(string(imageFormat))), // Required
			ImportManifestUrl: aws.String(getManiUrlStr),                        // Required
		},
		Volume: &ec2.VolumeDetail{ // Required
			Size: aws.Int64(toGigs(imageSize)), // Required
		},
	}
	task, err := ec2svc.ImportVolume(volparams)
	if err != nil {
		return "", err
	}
	log.WithFields(log.Fields{"task": *task}).Debug("Import task result")

	taskInput := &ec2.DescribeConversionTasksInput{
		ConversionTaskIds: []*string{task.ConversionTask.ConversionTaskId},
	}

	log.Debug("Waiting for task")
	err = ec2svc.WaitUntilConversionTaskCompleted(taskInput)

	if err != nil {
		return "", err
	}

	log.Debug("Task done")
	// hopefully successful!
	convTaskOutput, err := ec2svc.DescribeConversionTasks(taskInput)

	if err != nil {
		return "", err
	}

	log.WithFields(log.Fields{"task": *convTaskOutput}).Debug("Convertion task result")

	if len(convTaskOutput.ConversionTasks) != 1 {
		return "", errors.New("Unexpected number of tasks", nil)
	}
	convTask := convTaskOutput.ConversionTasks[0]

	if convTask.ImportVolume == nil {
		return "", errors.New("No volume information", nil)
	}

	return *convTask.ImportVolume.Volume.Id, nil

}

func uploadFileToAws(s3svc *s3.S3, file string, fileSize int64, bucket, path string) error {
	reader, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer reader.Close()
	return uploadToAws(s3svc, reader, fileSize, bucket, path)
}

func uploadToAws(s3svc *s3.S3, body io.ReadSeeker, size int64, bucket, path string) error {

	// upload
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket), // required
		Key:           aws.String(path),   // required
		ACL:           aws.String("private"),
		Body:          body,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String("application/octet-stream"),
	}

	_, err := s3svc.PutObject(params)

	if err != nil {
		return err
	}
	return nil
}

func toGigs(i int64) int64 {
	return 1 + (i >> 30)
}

func createBucket(s3svc *s3.S3, bucketName string) error {

	log.WithFields(log.Fields{"name": bucketName}).Debug("Creating Bucket ")

	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName), // Required
	}
	_, err := s3svc.CreateBucket(params)

	if err != nil {
		return err
	}

	return nil
}

func deleteBucket(s3svc *s3.S3, bucketName string) error {
	log.WithFields(log.Fields{"name": bucketName}).Debug("Deleting Bucket ")
	//first delete objects
	listObjectParams := &s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
	}
	objects, err := s3svc.ListObjects(listObjectParams)
	if err != nil {
		return err
	}
	for _, object := range objects.Contents {
		deleteObjectParams := &s3.DeleteObjectInput{
			Bucket: aws.String(bucketName),
			Key:    object.Key,
		}
		_, err := s3svc.DeleteObject(deleteObjectParams)
		if err != nil {
			return err
		}
	}

	params := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName), // Required
	}
	_, err = s3svc.DeleteBucket(params)
	return err
}

func deleteSnapshot(e2svc *ec2.EC2, snapshotId string) error {
	param := &ec2.DeleteSnapshotInput{
		SnapshotId: aws.String(snapshotId),
	}
	_, err := e2svc.DeleteSnapshot(param)
	return err
}

func deleteVolume(e2svc *ec2.EC2, volumeId string) error {
	param := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(volumeId),
	}
	_, err := e2svc.DeleteVolume(param)
	return err
}

type manifest struct {
	XMLName xml.Name `xml:"manifest"`

	Version         string   `xml:"version"`
	FileFormat      string   `xml:"file-format"`
	Importer        importer `xml:"importer"`
	SelfDestructUrl string   `xml:"self-destruct-url"`

	ImportSpec importSpec `xml:"import"`
}

type importer struct {
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Release string `xml:"release"`
}

type importSpec struct {
	Size       int64 `xml:"size"`
	VolumeSize int64 `xml:"volume-size"`
	Parts      parts `xml:"parts"`
}
type parts struct {
	Count int    `xml:"count,attr"`
	Parts []part `xml:"part"`
}

type part struct {
	Index     int       `xml:"index,attr"`
	ByteRange byteRange `xml:"byte-range"`
	Key       string    `xml:"key"`
	HeadUrl   string    `xml:"head-url"`
	GetUrl    string    `xml:"get-url"`
	DeleteUrl string    `xml:"delete-url"`
}
type byteRange struct {
	Start int64 `xml:"start,attr"`
	End   int64 `xml:"end,attr"`
}
