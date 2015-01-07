# aws-sts

Simple wrapper for the aws CLI tools to support authentication with MFA tokens.

## About

`aws-sts` uses "normal" AWS credentials and a provided MFA token to get STS tokens. These tokens are cached on disk so you are only asked for a new MFA token when the STS token expired.

Currently the default (and hardcoded) TTL for the STS token is 3600 seconds (so you will be asked for a new MFA token at most once an hour).

## Setup

`aws-sts` relies on the official AWS command lines tools which are best installed with `pip install awscli`.

`aws-sts` reads your original access key and secret from a JSON file containing these fields:

	{
		"aws_access_key_id": "YOUR_KEY",
		"aws_secret_access_key": "YOUR_SECRET",
		"aws_default_region": "DEFAULT_REGION"
	}

The location of that JSON file must be specified using the ENV variable `AWS_CREDENTIALS_PATH` 

You must also register one MFA device (currently the first mfa device found is used).

## Usage

### direct

		AWS_CREDENTIALS_PATH=/path/to/config.json aws-sts

### alias
Alternatively you can also create aliases for `aws-sts`

		alias aws="AWS_CREDENTIALS_PATH=/path/to/default.json aws-sts"
		alias aws-private="AWS_CREDENTIALS_PATH=/path/to/private.json aws-sts"
