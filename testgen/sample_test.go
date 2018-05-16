package testgen

import (
	"encoding/json"
	fmt "fmt"
	"reflect"
	"testing"

	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
)

func TestCopyIn(t *testing.T) {
	dst := TestGen{}
	src := TestGen{}
	src.Name = "foo"
	src.Db = 3.14159
	src.S64 = 64
	src.OuterEn = OuterEnum_OUTER2
	src.InnerEn = TestGen_INNER3
	src.InnerMsg = &TestGen_InnerMessage{
		Url: "myurl",
		Id:  1,
	}
	src.InnerMsgNonnull = TestGen_InnerMessage{
		Url: "myurl2",
		Id:  2,
	}
	src.IncludeMsg = &IncludeMessage{
		Name: "include message name",
		Id:   3,
		NestedMsg: &NestedMessage{
			Name: "nested name",
		},
	}
	src.IncludeMsgNonnull = IncludeMessage{
		Name: "include message name nonnull",
		Id:   4,
		NestedMsg: &NestedMessage{
			Name: "nested name2",
		},
	}
	src.IncludeFields = &IncludeFields{
		Name: "include fields name",
	}
	src.IncludeFieldsNonnull = IncludeFields{
		Name: "include fields name nonnull",
	}
	src.Loc = &edgeproto.Loc{
		Lat:  1.1,
		Long: 1.2,
	}
	src.LocNonnull = edgeproto.Loc{
		Lat:  2.1,
		Long: 2.2,
	}
	src.RepeatedInt = make([]int64, 10)
	for ii := 0; ii < 10; ii++ {
		src.RepeatedInt[ii] = int64(ii)
	}
	src.Ip = make([]byte, 4)
	for ii := 0; ii < 4; ii++ {
		src.Ip[ii] = byte(ii)
	}
	src.Names = make([]string, 5)
	for ii := 0; ii < 4; ii++ {
		src.Names[ii] = fmt.Sprintf("name %d", ii)
	}
	src.RepeatedMsg = make([]*IncludeMessage, 4)
	for ii := 0; ii < 4; ii++ {
		src.RepeatedMsg[ii] = &IncludeMessage{
			Name: "include message name",
			Id:   3,
			NestedMsg: &NestedMessage{
				Name: "nested name",
			},
		}
	}
	src.RepeatedMsgNonnull = make([]IncludeMessage, 4)
	for ii := 0; ii < 4; ii++ {
		src.RepeatedMsgNonnull[ii] = IncludeMessage{
			Name: "include message name",
			Id:   3,
			NestedMsg: &NestedMessage{
				Name: "nested name",
			},
		}
	}
	src.RepeatedFields = make([]*IncludeFields, 5)
	for ii := 0; ii < 5; ii++ {
		src.RepeatedFields[ii] = &IncludeFields{
			Name: "include fields name",
		}
	}
	src.RepeatedFieldsNonnull = make([]IncludeFields, 5)
	for ii := 0; ii < 5; ii++ {
		src.RepeatedFieldsNonnull[ii] = IncludeFields{
			Name: "include fields name",
		}
	}
	src.RepeatedInnerMsg = make([]*TestGen_InnerMessage, 3)
	for ii := 0; ii < 3; ii++ {
		src.RepeatedInnerMsg[ii] = &TestGen_InnerMessage{
			Url: "myurl2",
			Id:  2,
		}
	}
	src.RepeatedInnerMsgNonnull = make([]TestGen_InnerMessage, 3)
	for ii := 0; ii < 3; ii++ {
		src.RepeatedInnerMsgNonnull[ii] = TestGen_InnerMessage{
			Url: "myurl2",
			Id:  2,
		}
	}
	src.RepeatedLoc = make([]*edgeproto.Loc, 4)
	for ii := 0; ii < 4; ii++ {
		src.RepeatedLoc[ii] = &edgeproto.Loc{
			Lat:  2.1,
			Long: 2.2,
		}
	}
	src.RepeatedLocNonnull = make([]edgeproto.Loc, 4)
	for ii := 0; ii < 4; ii++ {
		src.RepeatedLocNonnull[ii] = edgeproto.Loc{
			Lat:  2.1,
			Long: 2.2,
		}
	}

	src.Fields = util.GrpcFieldsNew()
	util.GrpcFieldsSetAll(src.Fields)
	dst.CopyInFields(&src)
	// clear fields so they're not figured into the equals call
	src.Fields = nil
	equal := reflect.DeepEqual(dst, src)
	assert.True(t, equal, "src and dst are equal")
	if !equal {
		srcJson, err := json.Marshal(src)
		assert.Nil(t, err, "marshal src")
		dstJson, err := json.Marshal(dst)
		assert.Nil(t, err, "marshal dst")
		t.Errorf("src is %s", string(srcJson))
		t.Errorf("dst is %s", string(dstJson))
	}
}
