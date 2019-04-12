package log

import (
	"context"
	"fmt"
	"strings"
	"testing"

	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/stretchr/testify/assert"
)

type yamltest struct {
	Lvl  DebugLevel
	Lvls []DebugLevel
}

func TestYaml(t *testing.T) {
	y := yamltest{
		Lvl:  DebugLevel_api,
		Lvls: []DebugLevel{DebugLevel_notify, DebugLevel_etcd},
	}
	out, err := yaml.Marshal(&y)
	assert.Nil(t, err, "marshal yaml")

	strout := string(out)
	fmt.Println(strout)
	assert.True(t, strings.Index(strout, "lvl: api") != -1, "has api")
	assert.True(t, strings.Index(strout, "lvls: [notify, etcd]") != -1, "has array")

	yin := yamltest{}
	err = yaml.Unmarshal(out, &yin)
	assert.Nil(t, err, "unmarshal yaml")
	assert.Equal(t, y, yin, "equal")
}

func TestBits(t *testing.T) {
	lvl := uint64(1) | uint64(1)<<10
	SetDebugLevel(lvl)
	assert.Equal(t, lvl, debugLevel)
	ClearDebugLevel(uint64(1))
	assert.Equal(t, uint64(1)<<10, debugLevel)
	ClearDebugLevel(uint64(1) << 10)
	assert.Equal(t, uint64(0), debugLevel)

	SetDebugLevelEnum(DebugLevel_api)
	assert.True(t, debugLevel&DebugLevelApi != 0)
	SetDebugLevelEnum(DebugLevel_notify)
	assert.True(t, debugLevel&DebugLevelNotify != 0)
	ClearDebugLevelEnum(DebugLevel_api)
	assert.True(t, debugLevel&DebugLevelApi == 0)
	ClearDebugLevelEnum(DebugLevel_notify)
	assert.True(t, debugLevel&DebugLevelNotify == 0)

	SetDebugLevels([]DebugLevel{DebugLevel_api, DebugLevel_etcd})
	assert.True(t, debugLevel&DebugLevelApi != 0)
	assert.True(t, debugLevel&DebugLevelEtcd != 0)
	ClearDebugLevels([]DebugLevel{DebugLevel_notify, DebugLevel_etcd})
	assert.True(t, debugLevel&DebugLevelApi != 0)
	assert.True(t, debugLevel&DebugLevelNotify == 0)
	assert.True(t, debugLevel&DebugLevelEtcd == 0)
	ClearDebugLevels([]DebugLevel{DebugLevel_api, DebugLevel_etcd})
	assert.Equal(t, uint64(0), debugLevel)

	api := &Api{}
	lvls := DebugLevels{}
	lvls.Levels = []DebugLevel{DebugLevel_etcd, DebugLevel_api}
	_, err := api.EnableDebugLevels(context.TODO(), &lvls)
	assert.Nil(t, err)
	assert.True(t, debugLevel&DebugLevelApi != 0)
	assert.True(t, debugLevel&DebugLevelEtcd != 0)
	assert.True(t, debugLevel&DebugLevelNotify == 0)
	lvlsout, err := api.ShowDebugLevels(context.TODO(), &lvls)
	assert.Nil(t, err)
	assert.Equal(t, &lvls, lvlsout)
	_, err = api.DisableDebugLevels(context.TODO(), &lvls)
	assert.Nil(t, err)
	assert.Equal(t, uint64(0), debugLevel)
	lvlsout, err = api.ShowDebugLevels(context.TODO(), &lvls)
	assert.Nil(t, err)
	lvls = DebugLevels{Levels: []DebugLevel{}}
	assert.Equal(t, &lvls, lvlsout)
}
