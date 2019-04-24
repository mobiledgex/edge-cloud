package process

import (
	"os/exec"
	"reflect"
)

type Process interface {
	// Get the name of the process
	GetName() string
	// Get the hostname of the process
	GetHostname() string
	// Get EnvVars
	GetEnvVars() map[string]string
	// Start the process
	StartLocal(logfile string, opts ...StartOp) error
	// Stop the process
	StopLocal()
	// Get the exe name of the process binary
	GetExeName() string
	// Get lookup args that can be used to find the local process using pgrep
	LookupArgs() string
}

type Common struct {
	Kind        string
	Name        string
	Hostname    string
	DockerImage string
	EnvVars     map[string]string
	cmd         *exec.Cmd
}

func (c *Common) GetName() string {
	return c.Name
}

func (c *Common) GetHostname() string {
	return c.Hostname
}

func (c *Common) GetEnvVars() map[string]string {
	return c.EnvVars
}

// options

type StartOptions struct {
	Debug        string
	RolesFile    string
	CleanStartup bool
}

type StartOp func(op *StartOptions)

func WithDebug(debug string) StartOp {
	return func(op *StartOptions) { op.Debug = debug }
}

func WithRolesFile(rolesfile string) StartOp {
	return func(op *StartOptions) { op.RolesFile = rolesfile }
}

func WithCleanStartup() StartOp {
	return func(op *StartOptions) { op.CleanStartup = true }
}

func (s *StartOptions) ApplyStartOptions(opts ...StartOp) {
	for _, fn := range opts {
		fn(s)
	}
}

func GetTypeString(p interface{}) string {
	t := reflect.TypeOf(p)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}
