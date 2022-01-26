package dmecommon

import (
	"sync"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
)

type EcnStatInfo struct {
	EcnBit					 uint64
	SampleDurationMs uint64
	NumCe            uint64
	NumPackets       uint64
	Strategy         string
	Bandwidth        float64
}

// SignalStrength is a strange key/hash input.
func GetEcnStatKey(appInstKey edgeproto.AppInstKey, deviceInfo *DeviceInfo, loc *dme.Loc, tileLength int) EcnStatKey {
	statKey := EcnStatKey{
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

// Used to find corresponding EcnStat
// Created using EcnStatInfo
type EcnStatKey struct {
	AppInstKey      edgeproto.AppInstKey
	DeviceCarrier   string
	LocationTile    string
	DataNetworkType string
	DeviceOs        string
	DeviceModel     string
	SignalStrength  uint64
}

type EcnStat struct {
	NumSessions uint64 // number of sessions that send stats
	Mux         sync.Mutex
	Changed     bool
	Info        EcnStatInfo
}

func NewEcnStat() *EcnStat {
	return new(EcnStat)
}

func (d *EcnStat) Update(info *EcnStatInfo) {
	d.Changed = true
	d.NumSessions++

	d.Info = *info
}
