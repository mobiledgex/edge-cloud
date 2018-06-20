package testgen

import "github.com/mobiledgex/edge-cloud/edgeproto"

func MakeFieldMap(fields []string) map[string]struct{} {
	return edgeproto.MakeFieldMap(fields)
}
