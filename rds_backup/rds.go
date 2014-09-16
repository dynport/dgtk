package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/dynport/gocli"
	"github.com/dynport/gocloud/aws/rds"
)

var (
	rdsClient *rds.Client = rds.NewFromEnv()
)

type RDSBase struct {
	InstanceId string `cli:"arg required desc='RDS instance ID to fetch snapshots for'"`
}

type listRDSSnapshots struct {
	RDSBase
}

func (act *listRDSSnapshots) Run() (e error) {
	resp, e := (&rds.DescribeDBSnapshots{DBInstanceIdentifier: act.InstanceId}).Execute(rdsClient)
	if e != nil {
		return e
	}
	snapshots := resp.DescribeDBSnapshotsResult.Snapshots
	log.Printf("found %d snapshots", len(snapshots))

	table := gocli.NewTable()
	for i := range snapshots {
		table.Add(
			snapshots[i].DBInstanceIdentifier,
			snapshots[i].DBSnapshotIdentifier,
			snapshots[i].Status,
			snapshots[i].AllocatedStorage,
			snapshots[i].Engine,
			snapshots[i].EngineVersion,
			snapshots[i].SnapshotCreateTime,
		)
	}
	fmt.Println(table)

	return nil
}

type backupRDSSnapshot struct {
	RDSBase

	User      string `cli:"opt -u --user desc='user used for connection (database name by default)'"`
	Password  string `cli:"opt -p --pwd desc='password used for connection'"`
	TargetDir string `cli:"opt -d --dir default=. desc='path to save dumps to'"`

	Database string `cli:"arg required desc='the database to backup'"`
}

func (act *backupRDSSnapshot) user() string {
	if act.User == "" {
		return act.Database
	}
	return act.User
}

func (act *backupRDSSnapshot) dbSGName() string {
	return "sg-" + act.InstanceId + "-backup"
}

func (act *backupRDSSnapshot) dbInstanceId() string {
	return act.InstanceId + "-backup"
}

func (act *backupRDSSnapshot) Run() (e error) {
	// Create temporary DB security group with this host's public IP.
	if e = act.createDbSG(); e != nil {
		return e
	}
	defer func() { // Delete temporary DB security group.
		log.Printf("deleting db security group")
		err := act.deleteDbSG()
		if e == nil {
			e = err
		}
	}()

	// Select snapshot.
	snapshot, e := act.selectLatestSnapshot()
	if e != nil {
		return e
	}
	log.Printf("last snapshot %q from %s", snapshot.DBSnapshotIdentifier, snapshot.SnapshotCreateTime)

	// Determine target path and stop if dump already available (prior to creating the instance).
	var filename string
	if filename, e = act.createTargetPath(snapshot); e != nil {
		return e
	}

	// Restore snapshot into new instance.
	var instance *rds.DBInstance
	if instance, e = act.restoreDBInstance(snapshot); e != nil {
		log.Printf("failed to restore db instance: %s", e)
		return e
	}
	defer func() {
		log.Printf("deleting db instance")
		err := act.deleteDBInstance()
		if e == nil {
			e = err
		}
	}()

	return act.dumpDatabase(instance.Engine, instance.Endpoint.Address, instance.Endpoint.Port, filename)
}

func (act *backupRDSSnapshot) createTargetPath(snapshot *rds.DBSnapshot) (path string, e error) {
	path = filepath.Join(act.TargetDir, act.InstanceId)
	if e = os.MkdirAll(path, 0777); e != nil {
		return "", e
	}

	path = filepath.Join(path, fmt.Sprintf("%s.%s.gz", act.Database, snapshot.SnapshotCreateTime.Format("20060102T1504")))
	// make sure file does not exist yet.
	_, e = os.Stat(path)
	switch {
	case os.IsNotExist(e):
		e = nil
	case e == nil:
		e = os.ErrExist
	}

	return path, e
}

func (act *backupRDSSnapshot) createDbSG() (e error) {
	sgname := act.dbSGName()
	// Create a db security group to access the database.
	_, e = (&rds.CreateDBSecurityGroup{
		DBSecurityGroupName:        sgname,
		DBSecurityGroupDescription: "temporary db security group to create offsite backup",
	}).Execute(rdsClient)
	if e != nil {
		return e
	}
	log.Printf("created db security group %s", sgname)

	public, e := publicIP()
	if e != nil {
		return e
	}

	_, e = (&rds.AuthorizeDBSecurityGroupIngress{
		DBSecurityGroupName: sgname,
		CIDRIP:              public + "/32",
	}).Execute(rdsClient)
	if e != nil {
		return e
	}
	log.Printf("authorized %q on db security group %s", public, act.dbSGName())
	return nil
}

func (act *backupRDSSnapshot) deleteDbSG() (e error) {
	return (&rds.DeleteDBSecurityGroup{DBSecurityGroupName: act.dbSGName()}).Execute(rdsClient)
}

func (act *backupRDSSnapshot) selectLatestSnapshot() (*rds.DBSnapshot, error) {
	descResp, e := (&rds.DescribeDBSnapshots{DBInstanceIdentifier: act.InstanceId}).Execute(rdsClient)
	if e != nil {
		return nil, e
	}
	snapshots := descResp.DescribeDBSnapshotsResult.Snapshots

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots for %q found!", act.InstanceId)
	}

	max := struct {
		i int
		t time.Time
	}{0, snapshots[0].SnapshotCreateTime}

	for i := range snapshots {
		if max.t.Before(snapshots[i].SnapshotCreateTime) {
			max.i = i
			max.t = snapshots[i].SnapshotCreateTime
		}
	}
	return snapshots[max.i], nil
}

func (act *backupRDSSnapshot) dumpDatabase(engine, address, port, filename string) (e error) {
	var cmd *exec.Cmd
	switch engine {
	case "mysql":
		cmd = exec.Command("mysqldump", "--host="+address, "--port="+port, "--user="+act.user(), "--password="+act.Password, act.Database)
	case "postgres":
		cmd = exec.Command("pg_dump", "--host="+address, "--port", port, "--username="+act.user(), act.Database)
		cmd.Env = append(cmd.Env, "PGPASSWORD="+act.Password)
	default:
		return fmt.Errorf("engine %q not supported yet", engine)
	}

	fh, e := os.Create(filename)
	if e != nil {
		return e
	}
	defer fh.Close()

	gzw := gzip.NewWriter(fh)
	defer gzw.Close()

	cmd.Stdout = gzw
	cmd.Stderr = os.Stdout

	return cmd.Run()
}

func (act *backupRDSSnapshot) restoreDBInstance(snapshot *rds.DBSnapshot) (instance *rds.DBInstance, e error) {
	_, e = (&rds.RestoreDBSnapshot{
		DBInstanceIdentifier: act.dbInstanceId(),
		DBSnapshotIdentifier: snapshot.DBSnapshotIdentifier,
		DBInstanceClass:      "db.t1.micro",
	}).Execute(rdsClient)
	if e != nil {
		return nil, e
	}

	if _, e = act.waitForDBInstance(instanceAvailable); e != nil {
		return nil, e
	}

	_, e = (&rds.ModifyDBInstance{
		DBInstanceIdentifier: act.dbInstanceId(),
		DBSecurityGroups:     []string{act.dbSGName()},
	}).Execute(rdsClient)
	if e != nil {
		return nil, e
	}

	if instance, e = act.waitForDBInstance(instanceAvailable); e != nil {
		return nil, e
	}

	log.Printf("Created instance: %q in status %q reachable via %s", instance.DBInstanceIdentifier, instance.DBInstanceStatus, instance.Endpoint.Address)
	return instance, nil
}

func (act *backupRDSSnapshot) waitForDBInstance(f func([]*rds.DBInstance) bool) (instance *rds.DBInstance, e error) {
	// TODO: Add timeout.
	for {
		var instances []*rds.DBInstance
		instanceResp, e := (&rds.DescribeDBInstances{DBInstanceIdentifier: act.dbInstanceId()}).Execute(rdsClient)
		if e != nil {
			if err, ok := e.(rds.Error); !ok || err.Code != "DBInstanceNotFound" {
				return nil, e
			}
		} else {
			instances = instanceResp.DescribeDBInstancesResult.Instances
		}

		if f(instances) {
			if len(instances) == 1 {
				return instances[0], nil
			}
			return nil, nil
		}

		log.Printf("sleeping for 5 more seconds")
		time.Sleep(5 * time.Second)
	}
}

func instanceAvailable(instances []*rds.DBInstance) bool {
	return len(instances) == 1 && instances[0].DBInstanceStatus == "available"
}

func instanceGone(instances []*rds.DBInstance) bool {
	return len(instances) == 0
}

func (act *backupRDSSnapshot) deleteDBInstance() (e error) {
	_, e = (&rds.DeleteDBInstance{
		DBInstanceIdentifier: act.dbInstanceId(),
		SkipFinalSnapshot:    true,
	}).Execute(rdsClient)
	if e != nil {
		return e
	}
	_, e = act.waitForDBInstance(instanceGone)
	return e
}

func publicIP() (ip string, e error) {
	resp, e := http.Get("http://jsonip.com")
	if e != nil {
		return "", e
	}
	defer resp.Body.Close()

	res := map[string]string{}
	if e = json.NewDecoder(resp.Body).Decode(&res); e != nil {
		return "", e
	}

	if ip, ok := res["ip"]; ok {
		return ip, nil
	}
	return "", fmt.Errorf("failed to retrieve public ip")
}
