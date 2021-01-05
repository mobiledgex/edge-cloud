package testutil

import (
	"context"
	"log"
	"strings"

	"github.com/mobiledgex/edge-cloud/edgeproto"
)

const TagExpectErr = "expecterr"

type Run struct {
	client Client
	ctx    context.Context
	Errs   []Err
	Mode   string
	Rc     *bool
}

type Err struct {
	Desc string
	Msg  string
}

func NewRun(client Client, ctx context.Context, mode string, rc *bool) *Run {
	r := Run{
		client: client,
		ctx:    ctx,
		Mode:   mode,
		Rc:     rc,
	}
	r.Errs = make([]Err, 0)
	return &r
}

func (r *Run) logErr(desc string, err error) {
	if err == nil {
		return
	}
	e := Err{
		Desc: desc,
		Msg:  err.Error(),
	}
	r.Errs = append(r.Errs, e)
}

func (r *Run) CheckErrs(api, tag string) {
	if tag == TagExpectErr {
		// comparing output, do not fail for api errors
		return
	}
	// should not be any errors
	for _, err := range r.Errs {
		if strings.HasPrefix(api, "show") {
			if strings.Contains(err.Msg, "Forbidden") {
				continue
			}
		}
		log.Printf("\"%s\" run %s failed: %s\n", api, err.Desc, err.Msg)
		*r.Rc = false
	}
}

func FilterStreamResults(in [][]edgeproto.Result) [][]edgeproto.Result {
	filtered := make([][]edgeproto.Result, 0)
	for _, results := range in {
		ress := FilterResults(results)
		if len(ress) > 0 {
			filtered = append(filtered, ress)
		}
	}
	return filtered
}

// Remove results with code 0. This lets us remove status update results
// from create AppInst/ClusterInst/Cloudlet which are non-deterministic.
func FilterResults(in []edgeproto.Result) []edgeproto.Result {
	filtered := make([]edgeproto.Result, 0)
	for _, res := range in {
		if res.Code == 0 {
			continue
		}
		filtered = append(filtered, res)
	}
	return filtered
}
