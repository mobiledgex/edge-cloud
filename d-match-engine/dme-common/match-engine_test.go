package dmecommon

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type insertTestParams struct {
	limit int
	in    []float64
	out   []float64
}

func TestInsertResult(t *testing.T) {
	tests := []*insertTestParams{&insertTestParams{
		limit: 3,
		in:    []float64{1},
		out:   []float64{1},
	}, &insertTestParams{
		limit: 3,
		in:    []float64{5, 3},
		out:   []float64{3, 5},
	}, &insertTestParams{
		limit: 3,
		in:    []float64{3, 5},
		out:   []float64{3, 5},
	}, &insertTestParams{
		limit: 3,
		in:    []float64{5, 3, 1},
		out:   []float64{1, 3, 5},
	}, &insertTestParams{
		limit: 3,
		in:    []float64{1, 4, 5, 3},
		out:   []float64{1, 3, 4},
	}, &insertTestParams{
		limit: 3,
		in:    []float64{1, 4, 5, 3, 8, 2, 10},
		out:   []float64{1, 2, 3},
	}, &insertTestParams{
		limit: 5,
		in:    []float64{10, 4, 3, 8, 5, 2, 1},
		out:   []float64{1, 2, 3, 4, 5},
	}, &insertTestParams{
		limit: 4,
		in:    []float64{10, 4, 3, 8, 5, 2, 1},
		out:   []float64{1, 2, 3, 4},
	}, &insertTestParams{
		limit: 3,
		in:    []float64{10, 4, 3, 8, 5, 2, 1},
		out:   []float64{1, 2, 3},
	}, &insertTestParams{
		limit: 2,
		in:    []float64{10, 4, 3, 8, 5, 2, 1},
		out:   []float64{1, 2},
	}, &insertTestParams{
		limit: 1,
		in:    []float64{10, 4, 3, 8, 5, 2, 1},
		out:   []float64{1},
	}, &insertTestParams{
		limit: 0,
		in:    []float64{10, 4, 3, 8, 5, 2, 1},
		out:   []float64{},
	}}
	for _, params := range tests {
		testInsertObj(t, params)
	}
}

func testInsertObj(t *testing.T, params *insertTestParams) {
	fmt.Printf("testing: %+v\n", params)
	s := searchAppInst{
		resultLimit: params.limit,
	}
	for _, val := range params.in {
		f := &foundAppInst{
			distance: val,
		}
		s.insertResult(f)
	}
	results := make([]float64, 0)
	for _, f := range s.results {
		results = append(results, f.distance)
	}
	require.Equal(t, params.out, results)
}
