package edgeproto

import (
	"encoding/json"
	fmt "fmt"

	distributed_match_engine "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/objstore"
	context "golang.org/x/net/context"
)

func CheckForHttpPorts(ctx context.Context, objStore objstore.KVStore) error {
	keystr := fmt.Sprintf("%s/", objstore.DbKeyPrefixString("AppInst"))
	cbErrs := make([]error, 0)
	err := objStore.List(keystr, func(key, val []byte, rev, modRev int64) error {
		var appInst AppInst
		err2 := json.Unmarshal(val, &appInst)
		if err2 != nil {
			log.SpanLog(ctx, log.DebugLevelUpgrade, "Cannot unmarshal key", "val", string(val), "err", err2, "appInst", appInst)
			cbErrs = append(cbErrs, err2)
			return nil
		}
		log.SpanLog(ctx, log.DebugLevelUpgrade, "Checking AppInst for invalid legacy http ports", "appInst", appInst)
		for _, mappedPort := range appInst.MappedPorts {
			match := false
			for _, protoVal := range distributed_match_engine.LProto_value {
				if int32(mappedPort.Proto) == protoVal {
					match = true
				}
			}
			if !match {
				log.SpanLog(ctx, log.DebugLevelUpgrade, "Invalid protocol found", "appInst", appInst, "AppPort", mappedPort)
				err3 := fmt.Errorf("Invalid protocol found: %d", mappedPort.Proto)
				cbErrs = append(cbErrs, err3)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(cbErrs) == 0 {
		return nil
	}
	return fmt.Errorf("Errors: %v", cbErrs)
}
