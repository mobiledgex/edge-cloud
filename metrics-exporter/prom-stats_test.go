package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/stretchr/testify/assert"
)

var testMetricSent = 0

var testPayloadData = map[string]string{
	promQCpuClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491286.389,
				"10.01"
			  ]
			}
		  ]
		}
	  }`,
	promQMemClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491347.686,
				"99.99"
			  ]
			}
		  ]
		}
	  }`,
	promQDiskClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491384.455,
				"50.0"
			  ]
			}
		  ]
		}
	  }`,
	promQSendBytesRateClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491412.872,
				"11111"
			  ]
			}
		  ]
		}
	  }`,
	promQRecvBytesRateClust: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {},
			  "value": [
				1549491412.872,
				"22222"
			  ]
			}
		  ]
		}
	  }`,

	promQCpuPod: `{
		"status": "success",
		"data": {
		  "resultType": "vector",
		  "result": [
			{
			  "metric": {
				"pod_name": "testPod1"
			  },
			  "value": [
				1549491454.802,
				"5.0"
			  ]
			}
			]
		  }
		  }`,
	promQMemPod: `{
		"status": "success",
		"data": {
  		"resultType": "vector",
  		"result": [
			{
	  		"metric": {
				"pod_name": "testPod1"
	  		},
	  		"value": [
				1549484450.932,
				"100000000"
	  		]
			}
  		]
		}
		}`,
	promQNetSendRate: `{
		"status": "success",
		"data": {
  		"resultType": "vector",
  		"result": [
			{
	  		"metric": {
				"pod_name": "testPod1"
	  		},
	  		"value": [
				1549484450.932,
				"111111"
	  		]
			}
  		]
		}
		}`,
	promQNetRecvRate: `{
		"status": "success",
		"data": {
  		"resultType": "vector",
  		"result": [
			{
	  		"metric": {
				"pod_name": "testPod1"
	  		},
	  		"value": [
				1549484450.932,
				"222222"
	  		]
			}
  		]
		}
		}`,
}

func testMetricSend(metric *edgeproto.Metric) {
	testMetricSent = 1
}

func getTestMetrics(addr string, query string) (*PromResp, error) {
	input := []byte(testPayloadData[query])
	promResp := &PromResp{}
	if err := json.Unmarshal(input, &promResp); err != nil {
		return nil, err
	}
	return promResp, nil
}

func TestPromStats(t *testing.T) {
	*clusterName = "testcluster"
	testAppKey := MetricAppInstKey{
		cluster:   *clusterName,
		developer: "",
	}
	testPromStats := NewPromStats("0.0.0.0:9090", time.Second*1, testMetricSend)
	err := testPromStats.CollectPromStats(getTestMetrics)
	assert.Nil(t, err, "Fill stats from json")
	testAppKey.pod = "testPod1"
	stat, found := testPromStats.appStatsMap[testAppKey]
	// Check PodStats
	assert.True(t, found, "Pod testPod1 is not found")
	if found {
		assert.Equal(t, float64(5.0), stat.cpu)
		assert.Equal(t, uint64(100000000), stat.mem)
		assert.Equal(t, float64(0), stat.disk)
		assert.Equal(t, uint64(111111), stat.netSend)
		assert.Equal(t, uint64(222222), stat.netRecv)
	}
	// Check ClusterStats
	assert.Equal(t, float64(10.01), testPromStats.clusterStat.cpu)
	assert.Equal(t, float64(99.99), testPromStats.clusterStat.mem)
	assert.Equal(t, float64(50.0), testPromStats.clusterStat.disk)
	assert.Equal(t, uint64(11111), testPromStats.clusterStat.netSend)
	assert.Equal(t, uint64(22222), testPromStats.clusterStat.netRecv)

	// Check callback is called
	ts, _ := types.TimestampProto(time.Now())
	assert.Equal(t, int(0), testMetricSent)
	testPromStats.send(ClusterStatToMetric(ts, testPromStats.clusterStat))
	assert.Equal(t, int(1), testMetricSent)
}
