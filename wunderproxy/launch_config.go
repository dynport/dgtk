package wunderproxy

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/dynport/gocloud/aws/s3"
)

var ErrorLaunchConfigNotFound = fmt.Errorf("launch configuration not found")

// The launch configuration contains all the information required to start a
// container with the required environment, including all the environment
// variables required (like credentials).
type LaunchConfig struct {
	ContainerConfig *ContainerConfig
	HostConfig      *HostConfig
	HealthCheckPath string
	Revision        string
	ForceReload     bool

	hash string
}

// Function to load the currently deployed container for the environment
// specified in the prefix. If history is empty the ErrorLaunchConfigNotFound
// error type will be returned.
func LoadCurrentLaunchConfig(s3c *s3.Client, bucket, prefix string) (*LaunchConfig, error) {
	lc := new(LaunchConfig)
	return lc, lc.load(s3c, bucket, prefix, "current.json")
}

// Function to load a given launch configuration from S3. This must be a public
// method to be able to load it when first deploying it (its not part of the
// history right then).
func LoadLaunchConfig(s3c *s3.Client, bucket, prefix, hash string) (*LaunchConfig, error) {
	lc := new(LaunchConfig)
	err := lc.load(s3c, bucket, prefix, fmt.Sprintf("container.%s.json", hash))
	if err != nil {
		return nil, err
	}

	if lc.hash != hash {
		return nil, fmt.Errorf("given hash %q doesn't match actual hash %q", hash, lc.hash)
	}

	return lc, nil
}

func (lc *LaunchConfig) load(s3c *s3.Client, bucket, prefix, key string) error {
	resp, err := s3c.Get(bucket, prefix+"/"+key)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	buf := &bytes.Buffer{}
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 200:
		// ignore
	case 404:
		return ErrorLaunchConfigNotFound
	default:
		return fmt.Errorf("failed to request container config with key %q from s3: [%d] %s", key, resp.StatusCode, buf.String())
	}

	rawHash := md5.Sum(buf.Bytes())
	lc.hash = hex.EncodeToString(rawHash[:])

	return json.Unmarshal(buf.Bytes(), lc)
}

func (lc *LaunchConfig) save(s3client *s3.Client, bucket, prefix string) error {
	buf := bytes.NewBuffer(nil)
	err := json.NewEncoder(buf).Encode(lc)
	if err != nil {
		return err
	}

	rawHash := md5.Sum(buf.Bytes())
	lc.hash = hex.EncodeToString(rawHash[:])
	key := fmt.Sprintf("%s/container.%s.json", prefix, lc.hash)
	return s3client.Put(bucket, key, buf.Bytes(), nil)
}
