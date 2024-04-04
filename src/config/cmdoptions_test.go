package config

import (
	"os"
	"testing"

	flags "github.com/jessevdk/go-flags"
	"github.com/stretchr/testify/assert"
)

// NewCmdOptions returns a new instance of CmdOptions with default values
func NewCmdOptions(args ...string) *Options {
	cmdOpts := new(Options)
	_, _ = flags.NewParser(cmdOpts, flags.PrintErrors).ParseArgs(args)
	return cmdOpts
}

func TestParseFail(t *testing.T) {
	tests := [][]string{
		{0: "go-test", "--unknown-option"},
		{0: "go-test", "-c", "client01", "-f", "foo"},
	}
	for _, d := range tests {
		os.Args = d
		_, err := New(nil)
		assert.Error(t, err)
	}
}

func TestParseSuccess(t *testing.T) {
	tests := [][]string{
		{0: "go-test", "--version"},
	}
	for _, d := range tests {
		os.Args = d
		_, err := New(nil)
		assert.NoError(t, err)
	}
}

func TestLogLevel(t *testing.T) {
	c := &Options{Logging: LoggingOpts{LogLevel: "debug"}}
	assert.True(t, c.Verbose())
	c = &Options{Logging: LoggingOpts{LogLevel: "info"}}
	assert.False(t, c.Verbose())
}

func TestNewCmdOptions(t *testing.T) {
	c := NewCmdOptions("-c", "config_unit_test", "--password=somestrong")
	assert.NotNil(t, c)
}

func TestConfig(t *testing.T) {
	os.Args = []string{0: "config_test", "--config=sample.config.yaml"}
	_, err := New(nil)
	assert.NoError(t, err)

	os.Args = []string{0: "config_test", "--unknown"}
	_, err = New(nil)
	assert.Error(t, err)

	os.Args = []string{0: "config_test"} // clientname arg is missing, but set PW3_CONFIG
	assert.NoError(t, os.Setenv("PW3_CONFIG", "postgresql://foo:baz@bar/test"))
	_, err = New(nil)
	assert.NoError(t, err)
}
