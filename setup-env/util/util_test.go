package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testObj struct {
	InnerTestObj
	Strval1 string
	Strval2 string
	Intval1 int
	Intval2 int
	Inner1  InnerTestObj
	Inner2  *InnerTestObj
	Strarr  []string
	Intarr1 []int
	Intarr2 []*int
	Objarr1 []InnerTestObj
	Objarr2 []*InnerTestObj
	Strmap  map[string]string
	Intmap  map[int]int
	Objmap1 map[string]InnerTestObj
	Objmap2 map[string]*InnerTestObj
}

type InnerTestObj struct {
	Strval1 string
	Strval2 string
	Intval1 int
	Intval2 int
}

func getTestObj() *testObj {
	inner := InnerTestObj{
		Strval1: "instrval1",
		Strval2: "instrval2",
		Intval1: 300,
		Intval2: 400,
	}
	innerPtr := inner
	innerPtr2 := inner
	innerPtr3 := inner
	int1Ptr := 1

	return &testObj{
		InnerTestObj: InnerTestObj{
			Strval1: "embedstr1",
			Strval2: "embedstr2",
			Intval1: 10,
			Intval2: 11,
		},
		Strval1: "vstrval1",
		Strval2: "vstrval2",
		Intval1: 100,
		Intval2: 200,
		Inner1:  inner,
		Inner2:  &innerPtr,
		Strarr:  []string{"val1", "val2"},
		Intarr1: []int{1, 2, 3},
		Intarr2: []*int{&int1Ptr},
		Objarr1: []InnerTestObj{inner, inner},
		Objarr2: []*InnerTestObj{&innerPtr2},
		Strmap:  map[string]string{"foo": "bar"},
		Intmap:  map[int]int{1: 1, 2: 2},
		Objmap1: map[string]InnerTestObj{"foo1": inner},
		Objmap2: map[string]*InnerTestObj{"foo2": &innerPtr3},
	}
}

func TestTransformer(t *testing.T) {
	{
		// test: strings
		in := getTestObj()
		out := getTestObj()
		transformer := NewTransformer()
		transformer.AddSetZeroTypeField(testObj{}, "Strval2")
		transformer.Apply(in)
		out.Strval2 = ""
		require.Equal(t, out, in)
	}
	{
		// test: objects
		in := getTestObj()
		out := getTestObj()
		transformer := NewTransformer()
		transformer.AddSetZeroType(InnerTestObj{})
		transformer.Apply(in)
		out.InnerTestObj = InnerTestObj{}
		out.Inner1 = InnerTestObj{}
		out.Inner2 = &InnerTestObj{}
		out.Objarr1 = []InnerTestObj{InnerTestObj{}, InnerTestObj{}}
		out.Objarr2 = []*InnerTestObj{&InnerTestObj{}}
		out.Objmap1 = map[string]InnerTestObj{"foo1": InnerTestObj{}}
		out.Objmap2 = map[string]*InnerTestObj{"foo2": &InnerTestObj{}}
		require.Equal(t, out, in)
	}
	{
		// test: pointer to object
		in := getTestObj()
		out := getTestObj()
		transformer := NewTransformer()
		transformer.AddSetZeroType(&InnerTestObj{})
		transformer.Apply(in)
		out.Inner2 = nil
		out.Objarr2 = []*InnerTestObj{nil}
		out.Objmap2 = map[string]*InnerTestObj{"foo2": nil}
		require.Equal(t, out, in)
	}
	{
		// test: inner strings
		in := getTestObj()
		out := getTestObj()
		transformer := NewTransformer()
		transformer.AddSetZeroTypeField(InnerTestObj{}, "Strval2")
		transformer.Apply(in)
		out.InnerTestObj.Strval2 = ""
		out.Inner1.Strval2 = ""
		out.Inner2.Strval2 = ""
		for ii := range out.Objarr1 {
			out.Objarr1[ii].Strval2 = ""
		}
		for ii := range out.Objarr2 {
			out.Objarr2[ii].Strval2 = ""
		}
		for k, v := range out.Objmap1 {
			v.Strval2 = ""
			out.Objmap1[k] = v
		}
		for _, v := range out.Objmap2 {
			v.Strval2 = ""
		}
		require.Equal(t, out, in)
	}
}
