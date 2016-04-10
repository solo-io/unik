package aws

import (
	"encoding/xml"
	"io/ioutil"
	"os"
	"path"
	"time"

	"bytes"
	"io"

	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/emc-advanced-dev/unik/pkg/types"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"

	"math/rand"

	unikos "github.com/emc-advanced-dev/unik/pkg/os"

	"errors"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const DiskImageRaw = "RAW"

func UploadFileToAws(s3svc *s3.S3, file, bucket, path string) error {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return nil
	}

	reader, err := os.Open(file)
	if err != nil {
		return nil
	}
	defer reader.Close()
	return UploadToAws(s3svc, reader, fileInfo.Size(), bucket, path)
}

func UploadToAws(s3svc *s3.S3, body io.ReadSeeker, size int64, bucket, path string) error {

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
func CreateDataVolume(s3svc *s3.S3, ec2svc *ec2.EC2, folder string, az string) (*string, error) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(dir)
	imgFile := path.Join(dir, "vol.img")

	// only one partition - creat msdos partition which is most supported
	partitioner := func(device string) unikos.Partitioner { return &unikos.MsDosPartioner{Device: device} }
	unikos.CreateVolumes(imgFile, []types.RawVolume{types.RawVolume{Path: folder}}, partitioner)
	log.WithFields(log.Fields{"imgFile": imgFile}).Debug("Created temp image")

	return CreateDataVolumeFromRawImage(s3svc, ec2svc, imgFile, az)

}
func CreateDataVolumeFromRawImage(s3svc *s3.S3, ec2svc *ec2.EC2, imgFile string, az string) (*string, error) {

	fileInfo, err := os.Stat(imgFile)
	if err != nil {
		return nil, err
	}

	// upload the image file to aws
	bucket := fmt.Sprintf("unik-tmp-%d", rand.Int63())

	if err := CreateBucket(s3svc, bucket); err != nil {
		return nil, err
	}
	defer DeleteBucket(s3svc, bucket)

	pathInBucket := "disk.img"

	log.Debug("Uploading image to aws")

	if err := UploadFileToAws(s3svc, imgFile, bucket, pathInBucket); err != nil {
		return nil, err
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
		return nil, err
	}

	headReq, _ := s3svc.HeadObjectRequest(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(pathInBucket),
	})

	headUrlStr, err := headReq.Presign(24 * time.Hour)
	if err != nil {
		return nil, err
	}

	deleteReq, _ := s3svc.DeleteObjectRequest(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(pathInBucket),
	})

	deleteUrlStr, err := deleteReq.Presign(24 * time.Hour)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	m := Manifest{
		Version:         "2010-11-15",
		FileFormat:      DiskImageRaw,
		Importer:        Importer{"unik", "1", "2016-04-01"},
		SelfDestructUrl: deleteManiUrlStr,
		ImportSpec: ImportSpec{
			Size:       fileInfo.Size(),
			VolumeSize: toGigs(fileInfo.Size()),
			Parts: Parts{
				Count: 1,
				Parts: []Part{
					Part{
						Index: 0,
						ByteRange: ByteRange{
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
		return nil, err
	}
	log.Debug("Uploading manifest")

	// upload manifest
	manifestBytes := buf.Bytes()
	err = UploadToAws(s3svc, bytes.NewReader(manifestBytes), int64(len(manifestBytes)), bucket, manifestName)
	if err != nil {
		return nil, err
	}

	getManiReq, _ := s3svc.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(manifestName),
	})
	getManiUrlStr, err := getManiReq.Presign(24 * time.Hour)
	if err != nil {
		return nil, err
	}

	log.Debug("Importing volume")

	// finally import the image
	volparams := &ec2.ImportVolumeInput{
		AvailabilityZone: aws.String(az), // Required
		Image: &ec2.DiskImageDetail{ // Required
			Bytes:             aws.Int64(fileInfo.Size()), // Required
			Format:            aws.String(DiskImageRaw),   // Required
			ImportManifestUrl: aws.String(getManiUrlStr),  // Required
		},
		Volume: &ec2.VolumeDetail{ // Required
			Size: aws.Int64(1), // Required
		},
	}
	task, err := ec2svc.ImportVolume(volparams)
	if err != nil {
		return nil, err
	}
	log.WithFields(log.Fields{"task": *task}).Debug("Import task result")

	taskInput := &ec2.DescribeConversionTasksInput{
		ConversionTaskIds: []*string{task.ConversionTask.ConversionTaskId},
	}

	log.Debug("Waiting for task")
	err = ec2svc.WaitUntilConversionTaskCompleted(taskInput)

	if err != nil {
		return nil, err
	}

	log.Debug("Task done")
	// hopefully successful!
	convTaskOutput, err := ec2svc.DescribeConversionTasks(taskInput)

	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{"task": *convTaskOutput}).Debug("Convertion task result")

	if len(convTaskOutput.ConversionTasks) != 1 {
		return nil, errors.New("Unexpected number of tasks")
	}
	convTask := convTaskOutput.ConversionTasks[0]

	if convTask.ImportVolume == nil {
		return nil, errors.New("No volume information")
	}

	return convTask.ImportVolume.Volume.Id, nil

}

func toGigs(i int64) int64 {
	return 1 + (i >> 20)
}

type Manifest struct {
	XMLName xml.Name `xml:"manifest"`

	Version         string   `xml:"version"`
	FileFormat      string   `xml:"file-format"`
	Importer        Importer `xml:"importer"`
	SelfDestructUrl string   `xml:"self-destruct-url"`

	ImportSpec ImportSpec `xml:"import"`
}

type Importer struct {
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Release string `xml:"release"`
}

type ImportSpec struct {
	Size       int64 `xml:"size"`
	VolumeSize int64 `xml:"volume-size"`
	Parts      Parts `xml:"parts"`
}
type Parts struct {
	Count int    `xml:"count,attr"`
	Parts []Part `xml:"part"`
}

type Part struct {
	Index     int       `xml:"index,attr"`
	ByteRange ByteRange `xml:"byte-range"`
	Key       string    `xml:"key"`
	HeadUrl   string    `xml:"head-url"`
	GetUrl    string    `xml:"get-url"`
	DeleteUrl string    `xml:"delete-url"`
}
type ByteRange struct {
	Start int64 `xml:"start,attr"`
	End   int64 `xml:"end,attr"`
}

func CreateBucket(s3svc *s3.S3, bucketName string) error {

	log.WithFields(log.Fields{"name": bucketName}).Debug("Creating Bucket ")

	params := &s3.CreateBucketInput{
		Bucket: aws.String(bucketName), // Required
		// CreateBucketConfiguration: &s3.CreateBucketConfiguration{
		//     LocationConstraint : aws.String("us-east-1"),
		// },
	}
	_, err := s3svc.CreateBucket(params)

	if err != nil {
		return err
	}

	return nil
}

func DeleteBucket(s3svc *s3.S3, bucketName string) error {

	log.WithFields(log.Fields{"name": bucketName}).Debug("Deleting Bucket ")

	params := &s3.DeleteBucketInput{
		Bucket: aws.String(bucketName), // Required
	}
	_, err := s3svc.DeleteBucket(params)

	if err != nil {
		return err
	}

	return nil
}
