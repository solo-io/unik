package main

import (
	unikaws "github.com/emc-advanced-dev/unik/pkg/providers/aws"
    
	log "github.com/Sirupsen/logrus"
    
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main() {
    
	log.SetLevel(log.DebugLevel)
    
	var awsSession = session.New()
	// var meta = ec2metadata.New(awsSession)

	// region, err := meta.Region()
    region := "us-west-1"
    // bucket := "unikilicious"
    az := "us-west-1b" 
    folder := "/Users/kohavy/Work/unik/cmd/volume-uploader"
    
    config :=  &aws.Config{Region: aws.String(region)}
    ec2svc := ec2.New(awsSession, config)
    s3svc := s3.New(awsSession, config)

    _, err := unikaws.CreateDataVolume(s3svc, ec2svc, folder, az)
    
    if err != nil {
        panic(err)
    }
    
}