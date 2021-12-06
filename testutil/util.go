package testutil

import (
	"log"
	"sort"
	"strings"

	"github.com/mobiledgex/edge-cloud/objstore"
)

//based on the api some errors will be converted to no error
func ignoreExpectedErrors(api string, key objstore.ObjKey, err error) error {
	if err == nil {
		return err
	}
	if api == "delete" {
		if strings.Contains(err.Error(), key.NotFoundError().Error()) {
			log.Printf("ignoring error on delete : %v\n", err)
			return nil
		}
	} else if api == "create" {
		if strings.Contains(err.Error(), key.ExistsError().Error()) {
			log.Printf("ignoring error on create : %v\n", err)
			return nil
		}
	}
	return err
}

func (s *DebugDataOut) Sort() {
	for ii := 0; ii < len(s.Requests); ii++ {
		sort.Slice(s.Requests[ii], func(i, j int) bool {
			// ignore name for sorting
			ikey := s.Requests[ii][i].Node
			ikey.Name = ""
			jkey := s.Requests[ii][j].Node
			jkey.Name = ""
			return ikey.GetKeyString() < jkey.GetKeyString()
		})
	}
}
