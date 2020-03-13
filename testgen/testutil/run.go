package testutil

import "context"

// duplicate of testutil.Run for sample generator checking.
type Run struct {
	client Client
	ctx    context.Context
	Mode   string
	Rc     *bool
}

func (r *Run) logErr(desc string, err error) {}
