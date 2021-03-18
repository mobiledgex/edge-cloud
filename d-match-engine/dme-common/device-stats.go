package dmecommon

import (
	"sync"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

func GetDeviceStatKey(appInstKey edgeproto.AppInstKey, deviceInfo *dme.DeviceInfo, carrier string, loc *dme.Loc, tileLength int) DeviceStatKey {
	return DeviceStatKey{
		AppInstKey:      appInstKey,
		DataNetworkType: deviceInfo.DataNetworkType,
		DeviceOs:        deviceInfo.DeviceOs,
		DeviceModel:     deviceInfo.DeviceModel,
		SignalStrength:  uint64(deviceInfo.SignalStrength),
		DeviceCarrier:   carrier,
		LocationTile:    GetLocationTileFromGpsLocation(loc, tileLength),
	}
}

// Used to find corresponding CustomStat
// Created using CustomStatInfo
type DeviceStatKey struct {
	AppInstKey      edgeproto.AppInstKey
	DeviceCarrier   string
	LocationTile    string
	DataNetworkType string
	DeviceOs        string
	DeviceModel     string
	SignalStrength  uint64
}

type DeviceStat struct {
	NumSessions uint64 // number of sessions that send stats
	Mux         sync.Mutex
	Changed     bool
}

func NewDeviceStat() *DeviceStat {
	return new(DeviceStat)
}

func (d *DeviceStat) Update() {
	d.NumSessions++
}
