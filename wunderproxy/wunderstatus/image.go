package main

import (
	"fmt"

	"github.com/dynport/dgtk/wunderproxy"
	"github.com/dynport/gocloud/aws/s3"
)

type currentImage struct {
	S3Bucket string `cli:"opt required -s --s3bucket desc='S3 bucket used to load wunderproxy configuration'"`
	S3Prefix string `cli:"opt required -p --s3prefix desc='Key prefix used in the S3 bucket'"`
}

func (cc *currentImage) Run() error {
	s3client := s3.NewFromEnv()
	cfg, err := wunderproxy.LoadCurrentLaunchConfig(s3client, cc.S3Bucket, cc.S3Prefix)
	if err != nil {
		return err
	}

	fmt.Println(cfg.ContainerConfig.Image)
	return nil
}
