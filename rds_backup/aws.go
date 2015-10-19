package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/rds"
)

func newClient() *rds.RDS {
	cfg := loadConfig()
	return rds.New(cfg)
}

func newCloudWatchClient() *cloudwatch.CloudWatch {
	return cloudwatch.New(loadConfig())
}

func loadConfig() *aws.Config {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-west-1"
	}
	creds := credentials.NewEnvCredentials()
	return aws.NewConfig().WithCredentials(creds).WithRegion(region)
}
