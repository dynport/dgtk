# RDS Offsite Backup

## Build

Building is possible using the go way (for further information see the [go website](http://golang.org)).


## Usage

Call the built binary with

	rds_backup snapshot backup <rds-instance-id> <database>

There are options to specify the user and password using the `-u` and `-p`
flags.


## Custom Policy

Having a dedicated user with credentials that are only allowed to execute the
actions required is advised to prevent misuse in case of leaked credentials.
The following IAM policies are required. Make sure to replace the
`account-id` and `<identifier>` placeholders to appropriate values. Changing
the region might be required, too.

	{
	  "Version": "2012-10-17",
	  "Statement": [
	    {
	      "Sid": "Stmt1410789578000",
	      "Effect": "Allow",
	      "Action": [
	        "rds:DescribeDBSnapshots"
	      ],
	      "Resource": [
	        "arn:aws:rds:eu-west-1:<account-id>:db:<identifier>",
	      ]
	    },
	    {
	      "Sid": "Stmt1410790425000",
	      "Effect": "Allow",
	      "Action": [
	        "rds:CreateDBSecurityGroup",
	        "rds:DeleteDBSecurityGroup",
	        "rds:AuthorizeDBSecurityGroupIngress",
	        "rds:ModifyDBInstance"
	      ],
	      "Resource": [
	        "arn:aws:rds:eu-west-1:<account-id>:secgrp:sg-<identifier>-backup",
	      ]
	    },
	    {
	      "Sid": "Stmt1410790425001",
	      "Effect": "Allow",
	      "Action": [
	        "rds:RestoreDBInstanceFromDBSnapshot",
	        "rds:DescribeDBInstances",
	        "rds:ModifyDBInstance",
	        "rds:DeleteDBInstance"
	      ],
	      "Resource": [
	        "arn:aws:rds:eu-west-1:<account-id>:db:<identifier>-backup",
	        "arn:aws:rds:eu-west-1:<account-id>:snapshot:rds:<identifier>-*",
	      ]
	    }
	  ]
	}

