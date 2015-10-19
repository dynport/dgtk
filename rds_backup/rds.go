package main

type Base struct {
	InstanceId string `cli:"arg required desc='RDS instance ID to fetch snapshots for'"`
}
