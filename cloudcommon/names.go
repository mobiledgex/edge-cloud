package cloudcommon

import (
	"encoding/json"
	"fmt"
	"os"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/testutil"
	"github.com/mobiledgex/edge-cloud/util"
	yaml "gopkg.in/yaml.v2"
)

var AppDNSRoot = "mobiledgex.net"

var OperatorGCP = "gcp"
var OperatorAzure = "azure"
var OperatorDeveloper = "developer"

var AllocatedIpDynamic = "dynamic"

var RootLBL7Port = 443

// OperatorDeveloper is a special value used by the public cloud based cloudlet
var operatorDeveloper = edgeproto.OperatorKey{Name: OperatorDeveloper}

//DefaultCloudletKey is a special value for the public cloud based default cloudlet for each app
// is for an appinst deployment maintained by the developer, not Mobiledgex
var DefaultCloudletKey = edgeproto.CloudletKey{OperatorKey: operatorDeveloper, Name: "default"}

var PostgresUserName = "controller"
var PostgresUserDb = "userdb"

// GetRootLBFQDN gets the global Load Balancer's Fully Qualified Domain Name
// for apps using "shared" IP access.
func GetRootLBFQDN(key *edgeproto.CloudletKey) string {
	loc := util.DNSSanitize(key.Name)
	oper := util.DNSSanitize(key.OperatorKey.Name)
	return fmt.Sprintf("%s.%s.%s", loc, oper, AppDNSRoot)
}

// GetAppFQDN gets the app-specific Load Balancer's Fully Qualified Domain Name
// for apps using "dedicated" IP access.
func GetAppFQDN(key *edgeproto.AppInstKey) string {
	loc := util.DNSSanitize(key.CloudletKey.Name)
	oper := util.DNSSanitize(key.CloudletKey.OperatorKey.Name)
	dev := util.DNSSanitize(key.AppKey.DeveloperKey.Name)
	app := util.DNSSanitize(key.AppKey.Name)
	ver := util.DNSSanitize(key.AppKey.Version)
	return fmt.Sprintf("%s%s%s.%s.%s.%s", dev, app, ver, loc, oper, AppDNSRoot)
}

// GetL7Path gets the L7 path for L7 access behind the "shared"
// global Load Balancer (reverse proxy). This only the path and
// does not include the FQDN and port.
func GetL7Path(key *edgeproto.AppInstKey, port *dme.AppPort) string {
	dev := util.DNSSanitize(key.AppKey.DeveloperKey.Name)
	app := util.DNSSanitize(key.AppKey.Name)
	ver := util.DNSSanitize(key.AppKey.Version)
	return fmt.Sprintf("%s/%s%s/http%d", dev, app, ver, port.InternalPort)
}

// For the DME and CRM that require a cloudlet key to be specified
// at startup, this function parses the string argument.
func ParseMyCloudletKey(standalone bool, keystr *string, mykey *edgeproto.CloudletKey) {
	if standalone && *keystr == "" {
		// Use fake cloudlet
		*mykey = testutil.CloudletData[0].Key
		bytes, _ := json.Marshal(mykey)
		*keystr = string(bytes)
		return
	}

	if *keystr == "" {
		log.FatalLog("cloudletKey not specified")
	}

	err := json.Unmarshal([]byte(*keystr), mykey)
	if err != nil {
		err = yaml.Unmarshal([]byte(*keystr), mykey)
	}
	if err != nil {
		log.FatalLog("Failed to parse cloudletKey", "err", err)
	}

	err = mykey.Validate()
	if err != nil {
		log.FatalLog("Invalid cloudletKey", "err", err)
	}
}

func SetNodeKey(hostname *string, nodeType edgeproto.NodeType, cloudletKey *edgeproto.CloudletKey, key *edgeproto.NodeKey) {
	if *hostname == "" {
		*hostname, _ = os.Hostname()
		if *hostname == "" {
			*hostname = "nohostname"
		}
	}
	key.Name = *hostname
	key.NodeType = nodeType
	key.CloudletKey = *cloudletKey
}
