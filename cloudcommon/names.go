package cloudcommon

import (
	"fmt"

	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
)

var AppDNSRoot = "mobiledgex.net"

func GetRootLBFQDN(key *edgeproto.CloudletKey) string {
	loc := util.DNSSanitize(key.Name)
	oper := util.DNSSanitize(key.OperatorKey.Name)
	return fmt.Sprintf("%s.%s.%s", loc, oper, AppDNSRoot)
}
