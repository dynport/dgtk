package main

import (
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
)

func notifyCloudWatch(namespace, filename string) error {
	stat, err := os.Stat(filename)
	if err != nil {
		return err
	}

	cl := newCloudWatchClient()
	_, err = cl.PutMetricData(&cloudwatch.PutMetricDataInput{
		Namespace: aws.String(namespace),
		MetricData: []*cloudwatch.MetricDatum{
			&cloudwatch.MetricDatum{
				MetricName: aws.String("OffsiteBackupSize"),
				Unit:       aws.String("Bytes"),
				Value:      aws.Float64(float64(stat.Size())),
			},
		},
	})
	return err
}
