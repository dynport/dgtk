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
	return rds.New(&aws.Config{Credentials: credentials.NewEnvCredentials(), Region: region})
}

type Base struct {
	InstanceId string `cli:"arg required desc='RDS instance ID to fetch snapshots for'"`
}
