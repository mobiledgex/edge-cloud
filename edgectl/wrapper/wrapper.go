package wrapper

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/mobiledgex/edge-cloud/cli"
)

func RunEdgectl(args []string, ops ...RunOp) ([]byte, error) {
	opts := runOptions{}
	opts.apply(ops)

	extra := []string{"--output-format", "json-compact"}
	args = append(extra, args...)
	if opts.debug {
		log.Printf("running: edgectl %s\n", strings.Join(args, " "))
	}
	cmd := exec.Command("edgectl", args...)
	return cmd.CombinedOutput()
}

func RunEdgectlObjs(args []string, in, out interface{}, ops ...RunOp) error {
	opts := runOptions{}
	opts.apply(ops)

	objArgs, err := cli.MarshalArgs(in, opts.ignore, nil)
	if err != nil {
		return err
	}
	args = append(args, objArgs...)

	byt, err := RunEdgectl(args, ops...)
	if err != nil {
		return fmt.Errorf("%v, %s", err, string(byt))
	}
	str := strings.TrimSpace(string(byt))
	if str != "" {
		err = json.Unmarshal(byt, out)
		if err != nil {
			return fmt.Errorf("error '%v' unmarshaling: %s", err, string(byt))
		}
	}
	return nil
}

type runOptions struct {
	ignore       []string
	tls          string
	addr         string
	outputStream bool
	debug        bool
}

type RunOp func(opts *runOptions)

func WithIgnore(ignore []string) RunOp {
	return func(opts *runOptions) { opts.ignore = ignore }
}

func WithDebug() RunOp {
	return func(opts *runOptions) { opts.debug = true }
}

func (o *runOptions) apply(ops []RunOp) {
	for _, op := range ops {
		op(o)
	}
}
