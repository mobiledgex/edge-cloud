package dmecommon

import (
	"fmt"
	"math/rand"
	"sort"
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

func BenchmarkSort(b *testing.B) {
	// run: go test -bench BenchmarkSort
	// Test insertion sort vs built-in sort.
	// When limiting the number of results to small counts,
	// insertion sort performs much better than sorting the whole
	// list and then returning the first limit entries.
	// Typically GetAppInstList will limit results to 3 or 5,
	// and SDKDemoApp will want 20 or so.
	// With these values, regardless of number of AppInsts,
	// insertion sort will always be better.
	// However, if we need to start returning a large portion
	// of the entire AppInst set, then golang sort will be better.
	/*
		100insert3-4         300000      5088 ns/op
		100qsort3-4          100000     10329 ns/op
		100insert10-4        200000      5784 ns/op
		100qsort10-4         100000     10736 ns/op
		100insert20-4        200000      7222 ns/op
		100qsort20-4         200000     10265 ns/op
		100insert50-4        200000      9466 ns/op
		100qsort50-4         100000     10609 ns/op
		100insert100-4       200000     10826 ns/op
		100qsort100-4        200000     10723 ns/op

		1000insert3-4         30000     46899 ns/op
		1000qsort3-4          10000    153611 ns/op
		1000insert10-4        30000     48858 ns/op
		1000qsort10-4         10000    143714 ns/op
		1000insert20-4        30000     56920 ns/op
		1000qsort20-4         10000    143167 ns/op
		1000insert50-4        20000     81616 ns/op
		1000qsort50-4         10000    161585 ns/op
		1000insert100-4       10000    123773 ns/op
		1000qsort100-4        10000    144616 ns/op
		1000insert500-4        5000    258728 ns/op
		1000qsort500-4        10000    146211 ns/op
		1000insert1000-4       5000    294014 ns/op
		1000qsort1000-4       10000    139908 ns/op

		10000insert3-4         3000    428461 ns/op
		10000qsort3-4          1000   2017775 ns/op
		10000insert10-4        3000    463775 ns/op
		10000qsort10-4         1000   2030383 ns/op
		10000insert20-4        3000    545257 ns/op
		10000qsort20-4         1000   1983412 ns/op
		10000insert50-4        2000    785578 ns/op
		10000qsort50-4         1000   2032283 ns/op
		10000insert100-4       1000   1330272 ns/op
		10000qsort100-4        1000   1942636 ns/op
		10000insert500-4        300   4221638 ns/op
		10000qsort500-4        1000   1984822 ns/op
		10000insert1000-4       200   7579347 ns/op
		10000qsort1000-4       1000   1916734 ns/op
	*/
	rand.Seed(42)
	floats := []float64{}
	for ii := 0; ii < 10000; ii++ {
		floats = append(floats, rand.Float64())
	}
	numAppInsts := []int{100, 1000, 10000}
	limits := []int{3, 10, 20, 50, 100, 500, 1000}
	for _, num := range numAppInsts {
		for _, limit := range limits {
			if limit > num {
				continue
			}
			b.Run(fmt.Sprintf("%dinsert%d", num, limit), func(b *testing.B) {
				for c := 0; c < b.N; c++ {
					s := searchAppInst{
						resultLimit: limit,
					}
					for n := 0; n < num; n++ {
						f := &foundAppInst{
							distance: floats[n],
						}
						s.insertResult(f)
					}
				}
			})
			b.Run(fmt.Sprintf("%dqsort%d", num, limit), func(b *testing.B) {
				for c := 0; c < b.N; c++ {
					all := []*foundAppInst{}
					for n := 0; n < num; n++ {
						f := &foundAppInst{
							distance: floats[n],
						}
						all = append(all, f)
					}
					sort.Slice(all, func(i, j int) bool {
						return all[i].distance < all[j].distance
					})
				}
			})
		}
	}
}
