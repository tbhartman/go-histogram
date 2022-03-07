package histogram_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"golang.org/x/exp/constraints"

	"github.com/stretchr/testify/assert"
	"github.com/tbhartman/go-histogram"
)

func ExampleCreateLinear() {
	hist := histogram.CreateLinear([]float32{0, 0, 1, 2, 8}, 4)
	fmt.Printf("Edges: %v\n", hist.Edges())
	histogram.Print(os.Stdout, hist, histogram.PrintOptions[float32]{
		Width:     20,
		Prefix:    "  ",
		Symbol:    '>',
		PadSymbol: '-',
	})
	// Output:
	// Edges: [0 2 4 6 8]
	//     8
	//   > 6: ------>              (1)
	//   > 4: >                    (0)
	//   > 2: ------>              (1)
	//   > 0: -------------------> (3)
}

func counts[T constraints.Float | constraints.Integer](h histogram.Histogram[T]) []int {
	var ret = make([]int, len(h))
	for i, b := range h {
		ret[i] = b.Count
	}
	return ret
}

func TestHistCalc(t *testing.T) {
	a := assert.New(t)

	h := histogram.CreateLinear([]float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 10)
	a.Equal([]float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, h.Edges())
	// a.Equal([][]float32{{0}, {1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9, 10}}, h.Bins())
	a.Equal([]int{1, 1, 1, 1, 1, 1, 1, 1, 1, 2}, counts(h))

	h = histogram.CreateLinear([]float32{0, 1, 2, 4, 5, 6, 7, 8, 9, 10}, 10)
	a.Equal([]float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, h.Edges())
	// a.Equal([][]float32{{0}, {1}, {2}, nil, {4}, {5}, {6}, {7}, {8}, {9, 10}}, h.Bins())
	a.Equal([]int{1, 1, 1, 0, 1, 1, 1, 1, 1, 2}, counts(h))

	h = histogram.CreateLinear([]float32{0, 1, 2, 4, 4.5, 5, 6, 7, 8, 9, 10}, 10)
	a.Equal([]float32{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, h.Edges())
	// a.Equal([][]float32{{0}, {1}, {2}, nil, {4, 4.5}, {5}, {6}, {7}, {8}, {9, 10}}, h.Bins())
	a.Equal([]int{1, 1, 1, 0, 2, 1, 1, 1, 1, 2}, counts(h))

	h = histogram.CreateLinear([]float32{0, 1, 2, 4, 4.5, 5, 6, 7, 8, 9, 10}, 5)
	a.Equal([]float32{0, 2, 4, 6, 8, 10}, h.Edges())
	// a.Equal([][]float32{{0, 1}, {2}, {4, 4.5, 5}, {6, 7}, {8, 9, 10}}, h.Bins())
	a.Equal([]int{2, 1, 3, 2, 3}, counts(h))

	h = histogram.CreateLinear([]float32{0, 1}, 5)
	a.Equal([]float32{0, 0.2, 0.4, 0.6, 0.8, 1}, h.Edges())
	// a.Equal([][]float32{{0}, nil, nil, nil, {1}}, h.Bins())
	a.Equal([]int{1, 0, 0, 0, 1}, counts(h))

	h2 := histogram.CreateLinear([]int{0, 1}, 5)
	a.Equal([]int{0, 1, 2}, h2.Edges())
	// a.Equal([][]int{{0}, {1}}, h2.Bins())
	a.Equal([]int{1, 1}, counts(h2))
}

func TestHistPrint(t *testing.T) {
	s := []int{
		10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10, 10,
		1000,
		20, 300, 1000, 2000, 3000, 3000, 3000, 10000, 9500,
	}
	opt := histogram.GetBytesOptions[int](30, true)
	h := histogram.CreateLinear(s, 4)
	var b strings.Builder
	histogram.Print(&b, h, opt)
	assert.Equal(t, strings.TrimLeft(`
       9.8 KiB
  >    7.3 KiB: --|                          S ( 2;   19.0 KiB)
  >    4.9 KiB: $                              ( 0;    0     B)
  >    2.4 KiB: ---|         S                 ( 3;    8.8 KiB)
  >   10     B: -------S---------------------| (25;    4.4 KiB)
`, "\n"), b.String(), "need to rebase?\n"+b.String())

	b.Reset()
	h = histogram.CreateLog(s, 4)
	histogram.Print(&b, h, opt)
	assert.Equal(t, strings.TrimLeft(`
       9.8 KiB
  >    1.7 KiB: --------|                    S ( 6;   29.8 KiB)
  >  316     B: --S|                           ( 2;    2.0 KiB)
  >   56     B: S|                             ( 1;  300     B)
  >   10     B: S----------------------------| (21;  220     B)
`, "\n"), b.String(), "need to rebase?\n"+b.String())

	s = append([]int{0}, s...)
	b.Reset()
	h = histogram.CreateLog(s, 4)
	histogram.Print(&b, h, opt)
	assert.Equal(t, strings.TrimLeft(`
       9.8 KiB
  >    1.7 KiB: --------|                    S ( 6;   29.8 KiB)
  >  316     B: --S|                           ( 2;    2.0 KiB)
  >   56     B: S|                             ( 1;  300     B)
  >    0     B: S----------------------------| (22;  220     B)
`, "\n"), b.String(), "need to rebase?\n"+b.String())
}
