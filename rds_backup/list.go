package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/dynport/gocli"
)

type list struct {
	Base
}

func (act *list) Run() error {
	resp, err := newClient().DescribeDBSnapshots(&rds.DescribeDBSnapshotsInput{DBInstanceIdentifier: &act.InstanceId})
	if err != nil {
		return err
	}
	snapshots := resp.DBSnapshots
	logger.Printf("found %d snapshots", len(snapshots))

	table := gocli.NewTable()
	for i := range snapshots {
		t := ""
		if v := snapshots[i].SnapshotCreateTime; v != nil {
			t = v.UTC().Format("2006-01-02 15:04:05")
		}
		table.Add(
			p2s(snapshots[i].DBInstanceIdentifier),
			p2s(snapshots[i].DBSnapshotIdentifier),
			p2s(snapshots[i].Status),
			p2i64(snapshots[i].AllocatedStorage),
			p2s(snapshots[i].Engine),
			p2s(snapshots[i].EngineVersion),
			t,
		)
	}
	fmt.Println(table)
	return nil
}
