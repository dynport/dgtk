package main

import (
	"fmt"
	"strings"

	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/dgtk/wunderproxy"
	"github.com/dynport/dgtk/wunderproxy/wunderproxy/Godeps/_workspace/src/github.com/dynport/gocloud/aws/s3"
)

type currentEnv struct {
	S3Bucket string `cli:"opt required -s --s3bucket desc='S3 bucket used to load wunderproxy configuration'"`
	S3Prefix string `cli:"opt required -p --s3prefix desc='Key prefix used in the S3 bucket'"`
}

func (cc *currentEnv) Run() error {
	s3client := s3.NewFromEnv()
	cfg, err := wunderproxy.LoadCurrentLaunchConfig(s3client, cc.S3Bucket, cc.S3Prefix)
	if err != nil {
		return err
	}

	fmt.Println(strings.Join(cfg.ContainerConfig.Env, "\n"))
	return nil
}
