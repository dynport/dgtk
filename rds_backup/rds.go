package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/rds"
)

func newClient() *rds.RDS {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "eu-west-1"
	}
	creds := credentials.NewEnvCredentials()
	cfg := aws.NewConfig().WithCredentials(creds).WithRegion(region)
	return rds.New(cfg)
}

type Base struct {
	InstanceId string `cli:"arg required desc='RDS instance ID to fetch snapshots for'"`
}
