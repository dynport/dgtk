package gocli

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParse(t *testing.T) {
	args := &Args{}
	args.String("-h")

	args.Parse([]string{"droplets", "create", "-h=some.host"})
	assert.Equal(t, args.Args, []string{"droplets", "create"})
	assert.Equal(t, args.Get("-h"), []string{"some.host"})

	args.Parse([]string{"droplets", "create", "-h", "some.host"})
	assert.Equal(t, args.Args, []string{"droplets", "create"})
	assert.Equal(t, args.Get("-h"), []string{"some.host"})
}

func TestParseBool(t *testing.T) {
	args := &Args{}
	args.Bool("--rack")

	e := args.Parse([]string{"droplets", "create", "--rack"})
	assert.Nil(t, e)
	assert.Equal(t, args.Args, []string{"droplets", "create"})
	assert.Equal(t, args.GetBool("--rack"), true)
}

func TestNotRegistered(t *testing.T) {
	args := &Args{}
	assert.Nil(t, args.Parse([]string{"droplets", "create"}))
	assert.NotNil(t, args.Parse([]string{"droplets", "create", "--rack"}))
}

func TestStringWithoutDefault(t *testing.T) {
	args := &Args{}
	args.RegisterString("--host", "host", true, "", "Docker Host to be used")
	args.Parse([]string{"a", "b"})
	_, e := args.GetString("--host")
	assert.NotNil(t, e)
}

func TestRegister(t *testing.T) {
	args := NewArgs(map[string]*Flag{
		"-h": {Type: STRING, Required: false, DefaultValue: "some.host", Description: "Some Description"},
	},
	)
	assert.NotNil(t, args)
	ty, _ := args.TypeOf("-h")
	assert.Equal(t, ty, "string")
}

func TestStringWithDefault(t *testing.T) {
	args := &Args{}
	args.RegisterString("--host", "host", false, "default.host", "Docker Host to be used")
	args.RegisterBool("--rack", "rack", false, false, "Use as rack application")
	args.Parse([]string{"a", "b"})
	v, e := args.GetString("--host")
	assert.Nil(t, e)
	assert.Equal(t, v, "default.host")
}

func TestUsage(t *testing.T) {
	args := &Args{}
	args.RegisterString("--host", "host", false, "default.host", "Docker Host to be used")
	args.RegisterBool("--rack", "rack", true, false, "Use as rack application")
	s := args.Usage()
	assert.NotNil(t, s)
	assert.Contains(t, s, "--rack")
}

func TestRegisterFlag(t *testing.T) {
	args := &Args{}
	args.RegisterFlag(&Flag{CliFlag: "--host", Type: STRING})
	args.RegisterFlag(&Flag{CliFlag: "--help", Type: STRING})
	args.RegisterFlag(&Flag{CliFlag: "--enabled", Type: BOOL})

	assert.Equal(t, len(args.lookup("--h")), 2)
	assert.Equal(t, len(args.lookup("--h")), 2)
	assert.Equal(t, len(args.lookup("--ho")), 1)
	assert.Equal(t, len(args.lookup("--host")), 1)
}

func TestRegisterInt(t *testing.T) {
	args := &Args{}
	args.RegisterInt("-i", "image", false, 10, "I id")
	args.RegisterInt("-a", "a", false, 30, "A id")
	args.Parse([]string{"-i", "20"})

	v, _ := args.GetInt("-i")
	assert.Equal(t, v, 20)

	v, _ = args.GetInt("-a")
	assert.Equal(t, v, 30)
}

func TestAttributesMap(t *testing.T) {
	args := NewArgs(FlagMap{
		"-v":        {Type: STRING, DefaultValue: "1.2.3", Key: "version"},
		"--host":    {Type: STRING},
		"--user":    {Type: STRING},
		"--enabled": {Type: BOOL, DefaultValue: ""},
		"--flush":   {Type: BOOL, DefaultValue: "true"},
	})
	args.Parse([]string{"-v", "1.2.1", "--host", "localhost"})
	m := args.AttributesMap()
	assert.Equal(t, len(m), 3)
	assert.Equal(t, m, map[string]string{
		"version": "1.2.1",
		"host":    "localhost",
		"flush":   "true",
	})
}

func TestAttributesMapWithMultipleArgs(t *testing.T) {
	args := NewArgs(FlagMap{
		"-H": {Type: STRING, Key: "hosts"},
	})
	args.Parse([]string{"-H", "host1", "-H", "host2"})
	m := args.AttributesMap()
	assert.Equal(t, m["hosts"], "host1,host2")
}

func TestRegisterBool(t *testing.T) {
	args := &Args{}
	args.RegisterBool("--disabled", "disabled", false, false, "Disabled")
	tp, _ := args.TypeOf("--disabled")
	assert.Equal(t, tp, "bool")
	e := args.Parse([]string{"--disabled"})
	assert.Nil(t, e)
	res := args.GetBool("--disabled")
	assert.Equal(t, res, true)
}

func TestRegisterArgs(t *testing.T) {
	args := &Args{}
	args.RegisterArgs("command host")
}
