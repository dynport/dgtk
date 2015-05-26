package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	logger = log.New(os.Stderr, "", 0)
	dbg    = log.New(debugStream(), "[DEBUG] ", log.Lshortfile)
)

func main() {
	if err := run(); err != nil {
		logger.Fatal(err)
	}
}

func debugStream() io.Writer {
	if os.Getenv("DEBUG") == "true" {
		return os.Stderr
	}
	return ioutil.Discard
}

func extractName(p string) string {
	return strings.TrimSuffix(path.Base(p), ".json")
}

type config struct {
	AWSAccessKeyId     string `json:"aws_access_key_id,omitempty"`
	AWSSecretAccessKey string `json:"aws_secret_access_key,omitempty"`
	AWSDefaultRegion   string `json:"aws_default_region,omitempty"`
	AWSKeyName         string `json:"aws_key_name,omitempty"`
	AWSDurationSeconds int    `json:"aws_duration_seconds,omitempty"`
	AWSSessionToken    string `json:"aws_session_token,omitempty"`
}

func (c *config) Env() []string {
	o := []string{
		"AWS_ACCESS_KEY_ID=" + c.AWSAccessKeyId,
		"AWS_SECRET_ACCESS_KEY=" + c.AWSSecretAccessKey,
	}
	if c.AWSSessionToken != "" {
		o = append(o, "AWS_SESSION_TOKEN="+c.AWSSessionToken)
	}
	if c.AWSDefaultRegion != "" {
		o = append(o, "AWS_DEFAULT_REGION="+c.AWSDefaultRegion)
	}
	return o
}

type stsToken struct {
	SecretAccessKey string    `json:"SecretAccessKey,omitempty"`
	SessionToken    string    `json:"SessionToken,omitempty"`
	Expiration      time.Time `json:"Expiration,omitempty"`
	AccessKeyId     string    `json:"AccessKeyId,omitempty"`
}

func (c *stsToken) Env() []string {
	o := []string{
		"AWS_ACCESS_KEY_ID=" + c.AccessKeyId,
		"AWS_SECRET_ACCESS_KEY=" + c.SecretAccessKey,
		"AWS_SESSION_TOKEN=" + c.SessionToken,
	}
	return o
}

func loadConfig() (*config, error) {
	p := os.Getenv("AWS_CREDENTIALS_PATH")
	if p == "" {
		return nil, errors.New("AWS_CREDENTIALS_PATH must be set")
	}

	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg *config

	return cfg, json.NewDecoder(f).Decode(&cfg)
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	sts, err := loadSessionToken(cfg)
	if err != nil {
		return err
	}

	env := append(append(os.Environ(), "AWS_DEFAULT_REGION="+cfg.AWSDefaultRegion), sts.Env()...)
	args := []string{}
	if len(os.Args) > 1 {
		args = os.Args[1:]
	}
	c := exec.Command("aws", args...)
	c.Env = env
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Stdin = os.Stdin
	return c.Run()
}

func readToken() (string, error) {
	for {
		fmt.Print("token: ")
		tok, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			return "", err
		}
		tok = strings.TrimSpace(tok)
		if validToken(tok) {
			return tok, nil
		}
		fmt.Print("token not valid!\n")
	}
}

func loadSessionToken(cfg *config) (sts *stsToken, err error) {
	p := "/tmp/sts/" + cfg.AWSAccessKeyId + ".json"
	dbg.Printf("using sts path %q", p)
	f, err := os.Open(p)
	if err == nil {
		defer f.Close()
		err = json.NewDecoder(f).Decode(&sts)
		if err == nil {
			diff := sts.Expiration.Sub(time.Now())
			if diff > 0 {
				dbg.Printf("sts token found and still valid for %s", diff)
				return sts, nil
			}
		} else {
			dbg.Printf("ERROR reading sts token: %s", err)
		}
	}
	devs, err := listMFADevices(cfg)
	if err != nil {
		return nil, err
	}

	if len(devs) < 1 {
		return nil, fmt.Errorf("expected to find 1 mfa device, found %d", len(devs))
	}

	dev := devs[0]

	tok, err := readToken()
	if err != nil {
		return nil, err
	}
	dbg.Printf("token is %q", tok)
	dbg.Printf("getting session token with token %q and id %q", tok, dev.SerialNumber)
	sts, err = getSessionToken(cfg, tok, dev.SerialNumber, 8*3600)
	if err != nil {
		return nil, err
	}
	if err := storeSTSToken(p, sts); err != nil {
		logger.Printf("ERROR caching sts token: %s", err)
	}
	return sts, nil
}

func storeSTSToken(p string, sts *stsToken) error {
	t := p + ".tmp"
	err := os.MkdirAll(path.Dir(t), 0755)
	if err != nil {
		return err
	}
	f, err := os.Create(t)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(sts)
	if err != nil {
		return err
	}
	return os.Rename(t, p)
}

func getSessionToken(cfg *config, token, serial string, duration int) (*stsToken, error) {
	c := exec.Command("aws", "sts", "get-session-token", "--token-code="+token, "--serial-number="+serial, "--duration-seconds="+strconv.Itoa(duration))
	c.Env = cfg.Env()
	b, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, b)
	}
	var rsp *struct{ Credentials *stsToken }
	err = json.Unmarshal(b, &rsp)
	if err != nil {
		return nil, err
	}
	return rsp.Credentials, nil
}

func validToken(s string) bool {
	if len(s) != 6 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

type mfaDevice struct {
	UserName     string    `json:"UserName"`
	SerialNumber string    `json:"SerialNumber"`
	EnableDate   time.Time `json:"EnableDate"`
}

func listMFADevices(cfg *config) ([]*mfaDevice, error) {
	dbg.Printf("listing mfa devices")
	c := exec.Command("aws", "iam", "list-mfa-devices")
	c.Env = cfg.Env()
	b, err := c.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, b)
	}
	rsp := struct{ MFADevices []*mfaDevice }{}
	err = json.Unmarshal(b, &rsp)
	if err != nil {
		return nil, err
	}
	dbg.Printf("found %d mfa devices", len(rsp.MFADevices))
	return rsp.MFADevices, nil
}
