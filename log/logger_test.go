package log

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mobiledgex/edge-cloud/protoc-gen-cmd/yaml"
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
