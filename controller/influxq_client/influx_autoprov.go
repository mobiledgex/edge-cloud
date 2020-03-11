package influxq

import (
	"context"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/influxdata/influxdb/client/v2"
	"github.com/mobiledgex/edge-cloud/cloudcommon"
	"github.com/mobiledgex/edge-cloud/cloudcommon/influxsup"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
)

func (q *InfluxQ) PushAutoProvCounts(ctx context.Context, msg *edgeproto.AutoProvCounts) error {
	if q.client == nil {
		return fmt.Errorf("no client")
	}
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  q.dbName,
		Precision: InfluxQPrecision,
	})
	if err != nil {
		return err
	}
	ts, err := types.TimestampFromProto(&msg.Timestamp)
	if err != nil {
		return err
	}
	for ii, _ := range msg.Counts {
		apCount := msg.Counts[ii]
		pt, err := AutoProvCountToPt(apCount, msg.DmeNodeName, ts)
		if err != nil {
			log.SpanLog(ctx, log.DebugLevelMetrics, "metric new point", "err", err)
			continue
		}
		bp.AddPoint(pt)
	}
	return q.client.Write(bp)
}

func AutoProvCountToPt(apCount *edgeproto.AutoProvCount, dmeid string, ts time.Time) (*client.Point, error) {
	tags := make(map[string]string)
	tags["apporg"] = apCount.AppKey.Organization
	tags["app"] = apCount.AppKey.Name
	tags["ver"] = apCount.AppKey.Version
	tags["cloudletorg"] = apCount.CloudletKey.Organization
	tags["cloudlet"] = apCount.CloudletKey.Name
	tags["dme-id"] = dmeid
	fields := make(map[string]interface{})
	fields["count"] = int64(apCount.Count)
	return client.NewPoint(cloudcommon.AutoProvMeasurement, tags, fields, ts)
}

func ParseAutoProvCount(cols []string, values []interface{}) (*edgeproto.AutoProvCount, string, time.Time, error) {
	var id string
	var ts time.Time
	ap := edgeproto.AutoProvCount{}
	for c, name := range cols {
		val := values[c]
		var err error
		switch name {
		case "apporg":
			ap.AppKey.Organization, err = influxsup.ConvString(val)
		case "app":
			ap.AppKey.Name, err = influxsup.ConvString(val)
		case "ver":
			ap.AppKey.Version, err = influxsup.ConvString(val)
		case "cloudletorg":
			ap.CloudletKey.Organization, err = influxsup.ConvString(val)
		case "cloudlet":
			ap.CloudletKey.Name, err = influxsup.ConvString(val)
		case "dme-id":
			id, err = influxsup.ConvString(val)
		case "count":
			var count int64
			count, err = influxsup.ConvInt(val)
			ap.Count = uint64(count)
		case "time":
			ts, err = influxsup.ConvTime(val)
		}
		if err != nil {
			err = fmt.Errorf("failed to parse field \"%s\", %v", name, err)
			return &ap, id, ts, err
		}
	}
	return &ap, id, ts, nil
}
