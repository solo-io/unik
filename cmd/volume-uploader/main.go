package main

import (
	"flag"

	unikaws "github.com/emc-advanced-dev/unik/pkg/providers/aws"

	log "github.com/Sirupsen/logrus"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {

	log.SetLevel(log.DebugLevel)

	region := flag.String("r", "us-west-1", "region")
	az := flag.String("a", "us-west-1b", "availability zone")
	imgFile := flag.String("f", nil, "Image file")

	flag.Parse()

	if imgFile == nil {
		log.Fatal("Must provide image file")
	}

	var awsSession = session.New()

	config := &aws.Config{Region: region}
	ec2svc := ec2.New(awsSession, config)
	s3svc := s3.New(awsSession, config)

	_, err := unikaws.createDataVolumeFromRawImage(s3svc, ec2svc, *imgFile, *az)

	if err != nil {
		panic(err)
	}

}
