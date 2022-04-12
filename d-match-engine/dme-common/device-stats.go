// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
