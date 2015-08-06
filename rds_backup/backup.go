package main

import (
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/service/rds"
)

const encryptionKey = "gaN5joJ0Niv3peed0asT0Yoim1yAd2bO"

type backup struct {
	Base

	User         string `cli:"opt -u --user desc='user used for connection (database name by default)'"`
	Password     string `cli:"opt -p --pwd desc='password used for connection'"`
	TargetDir    string `cli:"opt -d --dir default=. desc='path to save dumps to'"`
	InstanceType string `cli:"opt -t --instance-type default=db.t1.micro desc='db instance type'"`
	Uncompressed bool   `cli:"opt --uncompressed desc='run dump uncompressed'"`

	NoEncryption bool `cli:"opt --no-encryption desc='do not encrypt the dump file'"`

	Database string   `cli:"arg required desc='the database to backup'"`
	Tables   []string `cli:"arg desc='list of tables to dump (all if not specified)'"`
}

func (act *backup) user() string {
	if act.User == "" {
		return act.Database
	}
	return act.User
}

func (act *backup) dbSGName() string {
	return "sg-" + act.InstanceId + "-backup"
}

func (act *backup) dbInstanceId() string {
	return act.InstanceId + "-backup"
}

func (act *backup) Run() (e error) {
	// Create temporary DB security group with this host's public IP.
	if e = act.createDbSG(); e != nil {
		return e
	}
	defer func() { // Delete temporary DB security group.
		logger.Printf("deleting db security group")
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
	logger.Printf("last snapshot %q from %s", p2s(snapshot.DBSnapshotIdentifier), snapshot.SnapshotCreateTime)

	if snapshot.SnapshotCreateTime.Before(time.Now().Add(-24 * time.Hour)) {
		return fmt.Errorf("latest snapshot older than 24 hours!")
	}

	// Restore snapshot into new instance.
	var instance *rds.DBInstance
	if instance, e = act.restoreDBInstance(snapshot); e != nil {
		logger.Printf("failed to restore db instance: %s", e)
		return e
	}
	defer func() {
		logger.Printf("deleting db instance")
		err := act.deleteDBInstance()
		if e == nil {
			e = err
		}
	}()

	var filename string
	if filename, e = act.createTargetPath(snapshot); e != nil {
		return e
	}

	for i := 0; i < 3; i++ {
		// Determine target path and stop if dump already available (prior to creating the instance).
		logger.Printf("dumping database, try %d", i+1)
		e = act.dumpDatabase(*instance.Engine, *instance.Endpoint.Address, *instance.Endpoint.Port, filename)
		if e != nil {
			logger.Printf("ERROR dumping database: step=%d %s", i+1, e)
		} else {
			return nil
		}
	}
	return e
}

func (act *backup) createTargetPath(snapshot *rds.DBSnapshot) (path string, e error) {
	path = filepath.Join(act.TargetDir, act.InstanceId)
	if e = os.MkdirAll(path, 0777); e != nil {
		return "", e
	}

	suffix := ".sql"
	if !act.Uncompressed {
		suffix += ".gz"
	}
	path = filepath.Join(path, fmt.Sprintf("%s.%s.%s", act.Database, snapshot.SnapshotCreateTime.Format("20060102T1504"), suffix))
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

func (act *backup) dumpDatabase(engine, address string, port int64, filename string) (e error) {
	defer benchmark("dump database to " + filename)()
	var cmd *exec.Cmd
	compressed := false
	portS := strconv.FormatInt(port, 10)
	switch engine {
	case "mysql":
		args := []string{"--host=" + address, "--port=" + portS, "--user=" + act.user(), "--password=" + act.Password}
		if !act.Uncompressed {
			args = append(args, "--compress")
		}
		args = append(args, act.Database)
		if act.Tables != nil && len(act.Tables) > 0 {
			args = append(args, act.Tables...)
		}
		cmd = exec.Command("mysqldump", args...)
	case "postgres":
		args := []string{"--host=" + address, "--port=" + portS, "--username=" + act.user()}
		if !act.Uncompressed {
			args = append(args, "--compress=6")
		}
		args = append(args, act.Database)
		for i := range act.Tables {
			args = append(args, "-t", act.Tables[i])
		}
		cmd = exec.Command("pg_dump", args...)
		cmd.Env = append(cmd.Env, "PGPASSWORD="+act.Password)
		compressed = true
	default:
		return fmt.Errorf("engine %q not supported yet", engine)
	}

	tmpName := filename + ".tmp"
	fh, e := os.OpenFile(tmpName, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0644)
	if e != nil {
		return fmt.Errorf("ERROR opening file %q: %s", tmpName, e)
	}
	defer deferredClose(fh, &e)

	var encWriter io.Writer
	if act.NoEncryption {
		encWriter = fh
	} else {
		block, err := aes.NewCipher([]byte(encryptionKey))
		if err != nil {
			return err
		}

		// If the key is unique for each ciphertext, then it's ok to use a zero
		// IV.
		var iv [aes.BlockSize]byte
		stream := cipher.NewOFB(block, iv[:])

		encWriter = &cipher.StreamWriter{S: stream, W: fh}
	}

	if compressed || act.Uncompressed {
		cmd.Stdout = encWriter
	} else {
		gzw := gzip.NewWriter(encWriter)
		defer deferredClose(gzw, &e)
		cmd.Stdout = gzw
	}

	cmd.Stderr = os.Stdout
	e = cmd.Run()
	if e != nil {
		_ = os.Remove(tmpName)
		return e
	}
	e = os.Rename(tmpName, filename)
	if e != nil {
		return fmt.Errorf("ERROR renaming file %q to %q: %s", tmpName, filename, e)
	}
	return nil
}

func (act *backup) restoreDBInstance(snapshot *rds.DBSnapshot) (instance *rds.DBInstance, err error) {
	defer benchmark("restoreDBInstance")()
	client := newClient()

	if _, err := client.RestoreDBInstanceFromDBSnapshot(&rds.RestoreDBInstanceFromDBSnapshotInput{
		DBInstanceIdentifier: s2p(act.dbInstanceId()),
		DBSnapshotIdentifier: snapshot.DBSnapshotIdentifier,
		DBInstanceClass:      &act.InstanceType,
	}); err != nil {
		return nil, fmt.Errorf("[restore instance] %s", err)
	}

	if _, err := act.waitForDBInstance(instanceAvailable); err != nil {
		return nil, fmt.Errorf("[waiting] %s", err)
	}

	if _, err := client.ModifyDBInstance(&rds.ModifyDBInstanceInput{
		DBInstanceIdentifier: s2p(act.dbInstanceId()),
		DBSecurityGroups:     []*string{s2p(act.dbSGName())},
	}); err != nil {
		return nil, fmt.Errorf("[modify] %s", err)
	}

	if instance, err = act.waitForDBInstance(instancePortAvailable); err != nil {
		return nil, fmt.Errorf("[waiting 2] %s", err)
	}

	logger.Printf("Created instance: %q in status %q reachable via %s", p2s(instance.DBInstanceIdentifier), p2s(instance.DBInstanceStatus), p2s(instance.Endpoint.Address))
	return instance, nil
}

func (act *backup) waitForDBInstance(f func([]*rds.DBInstance) bool) (instance *rds.DBInstance, e error) {
	// TODO: Add timeout.
	client := newClient()
	for {
		var instances []*rds.DBInstance

		instanceResp, err := client.DescribeDBInstances(&rds.DescribeDBInstancesInput{DBInstanceIdentifier: s2p(act.dbInstanceId())})
		if err != nil {
			return nil, err
		} else {
			instances = instanceResp.DBInstances
		}

		if f(instances) {
			if len(instances) == 1 {
				return instances[0], nil
			}
			return nil, nil // instances is empty when waiting for termination
		}

		dbg.Printf("sleeping for 5 more seconds")
		time.Sleep(5 * time.Second)
	}
}

func (act *backup) createDbSG() (e error) {
	sgname := act.dbSGName()
	desc := "temporary db security group to create offsite backup"
	// Create a db security group to access the database.

	client := newClient()
	_, err := client.CreateDBSecurityGroup(&rds.CreateDBSecurityGroupInput{
		DBSecurityGroupName:        &sgname,
		DBSecurityGroupDescription: &desc,
	})
	if err != nil {
		return err
	}
	logger.Printf("created db security group %s", sgname)

	public, e := publicIP()
	if e != nil {
		return e
	}

	_, err = client.AuthorizeDBSecurityGroupIngress(&rds.AuthorizeDBSecurityGroupIngressInput{
		DBSecurityGroupName: &sgname,
		CIDRIP:              s2p(public + "/32"),
	})
	if err != nil {
		return err
	}
	logger.Printf("authorized %q on db security group %s", public, act.dbSGName())
	return nil
}

func (act *backup) deleteDbSG() error {
	name := act.dbSGName()
	_, err := newClient().DeleteDBSecurityGroup(&rds.DeleteDBSecurityGroupInput{DBSecurityGroupName: &name})
	return err
}

func (act *backup) selectLatestSnapshot() (*rds.DBSnapshot, error) {
	descResp, e := newClient().DescribeDBSnapshots(&rds.DescribeDBSnapshotsInput{DBInstanceIdentifier: &act.InstanceId})
	if e != nil {
		return nil, e
	}
	snapshots := descResp.DBSnapshots

	if len(snapshots) == 0 {
		return nil, fmt.Errorf("no snapshots for %q found!", act.InstanceId)
	}

	var snapshot *rds.DBSnapshot

	for _, current := range snapshots {
		if current.SnapshotCreateTime == nil {
			continue
		}
		if snapshot == nil {
			snapshot = current
		} else if current.SnapshotCreateTime.After(*snapshot.SnapshotCreateTime) {
			snapshot = current
		}
	}
	if snapshot == nil {
		return nil, fmt.Errorf("no snapshot with timestamp found for %q", act.InstanceId)
	}
	return snapshot, nil
}

func instanceAvailable(instances []*rds.DBInstance) bool {
	return len(instances) == 1 && p2s(instances[0].DBInstanceStatus) == "available"
}

func instancePortAvailable(instances []*rds.DBInstance) bool {
	if len(instances) != 1 {
		return false
	}
	ins := instances[0]
	l, e := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", *ins.Endpoint.Address, *ins.Endpoint.Port), 1*time.Second)
	if e != nil {
		return false
	}
	defer l.Close()
	return true
}

func instanceGone(instances []*rds.DBInstance) bool {
	return len(instances) == 0
}

func (act *backup) deleteDBInstance() error {
	defer benchmark("deleteDBInstance")()
	client := newClient()
	skip := true
	_, err := client.DeleteDBInstance(&rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: s2p(act.dbInstanceId()),
		SkipFinalSnapshot:    &skip,
	})
	if err != nil {
		return err
	}

	_, err = act.waitForDBInstance(instanceGone)
	return err
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

func deferredClose(c io.Closer, e *error) {
	if err := c.Close(); err != nil && *e == nil {
		*e = err
	}
}
