package dmecommon

import (
	"sync"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type DeviceInfo struct {
	DeviceInfoStatic  *dme.DeviceInfoStatic
	DeviceInfoDynamic *dme.DeviceInfoDynamic
}

func GetDeviceStatKey(appInstKey edgeproto.AppInstKey, deviceInfo *DeviceInfo, loc *dme.Loc, tileLength int) DeviceStatKey {
	statKey := DeviceStatKey{
		AppInstKey:   appInstKey,
		LocationTile: GetLocationTileFromGpsLocation(loc, tileLength),
	}
	if deviceInfo.DeviceInfoStatic != nil {
		statKey.DeviceOs = deviceInfo.DeviceInfoStatic.DeviceOs
		statKey.DeviceModel = deviceInfo.DeviceInfoStatic.DeviceModel
	}
	if deviceInfo.DeviceInfoDynamic != nil {
		statKey.DeviceCarrier = deviceInfo.DeviceInfoDynamic.CarrierName
		statKey.DataNetworkType = deviceInfo.DeviceInfoDynamic.DataNetworkType
		statKey.SignalStrength = uint64(deviceInfo.DeviceInfoDynamic.SignalStrength)
	}
	return statKey
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
	d.Changed = true
	d.NumSessions++
}
