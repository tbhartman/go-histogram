package histogram

import (
	"constraints"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
)

type number interface {
	constraints.Float | constraints.Integer
}

func sum[T number](s []T) T {
	var ret T
	for _, v := range s {
		ret += v
	}
	return ret
}

func apply[T number](s []T, f func(a, b T) T) T {
	if len(s) == 0 {
		return 0
	}
	var ret T = s[0]
	for _, v := range s[1:] {
		ret = f(ret, v)
	}
	return ret
}

func min[T number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func max[T number](a, b T) T {
	if a > b {
		return a
	}
	return b
}

func minSlice[T number](s []T) T { return apply(s, min[T]) }
func maxSlice[T number](s []T) T { return apply(s, max[T]) }

// Bucket is one element of a Histogram
type Bucket[T constraints.Float | constraints.Integer] struct {
	_     [0]func() // Bucket is not comparable
	Low   T         // Low is the low edge of this Bucket
	High  T         // High is the high edge of this Bucket
	Min   T         // Min is the minimum of values in this Bucket, if Count > 0
	Max   T         // Max is the maximum of values in this Bucket, if Count > 0
	Count int       // Count is the number of values in this bucket
	Sum   T         // Sum is the summation of all values in this bucket
}

func (b *Bucket[T]) add(value T) {
	b.Count++
	b.Sum += value
	if b.Count == 1 {
		b.Min = value
		b.Max = value
	} else if b.Min > value {
		b.Min = value
	} else if b.Max < value {
		b.Max = value
	}
}

// Histogram is an ordered slice of Buckets
type Histogram[T constraints.Float | constraints.Integer] []Bucket[T]

// Edges returns a slice of edge values of the Buckets
func (h Histogram[T]) Edges() []T {
	var ret []T
	for i, b := range h {
		if i == 0 {
			ret = append(ret, b.Low)
		}
		ret = append(ret, b.High)
	}
	return ret
}

// CreateLog creates a Histogram on a log scale
func CreateLog[T constraints.Float | constraints.Integer](slice []T, binCount int) Histogram[T] {
	// find minimum that is greater than zero
	var nonzeroMin T
	var absMinFound bool
	var absMin T
	for _, s := range slice {
		if s > 0 {
			if nonzeroMin == 0 || s < nonzeroMin {
				nonzeroMin = s
			}
		}
		if !absMinFound || s < absMin {
			absMinFound = true
			absMin = s
		}
	}
	h := create(slice, binCount,
		func(in T) float64 {
			if in < nonzeroMin {
				in = nonzeroMin
			}
			return math.Log10(float64(in))
		},
		func(in float64) T { return roundIfInt[T](math.Pow(10, in)) },
	)
	if len(h) > 0 {
		h[0].Low = absMin
	}
	return h
}

// CreateLinear creates a Histogram on a linear scale
func CreateLinear[T constraints.Float | constraints.Integer](slice []T, binCount int) Histogram[T] {
	return create(slice, binCount, nil, nil)
}

func isInt[T constraints.Float | constraints.Integer]() bool {
	return T(1)/T(3) == 0
}
func roundIfInt[T constraints.Float | constraints.Integer](value float64) T {
	if !isInt[T]() {
		return T(value)
	}
	return T(math.Round(value))
}

// beforeCalc, afterCalc must preserve sorting
func create[T constraints.Float | constraints.Integer](inSlice []T, binCount int, beforeCalc func(T) float64, afterCalc func(float64) T) (ret Histogram[T]) {
	if len(inSlice) == 0 {
		return
	}
	var min = minSlice(inSlice)
	var max = maxSlice(inSlice)

	// if an integer, reduce binCount difference in min/max is less than binCount
	if isInt[T]() && int(max-min) < binCount {
		binCount = int(max-min) + 1
		max++
	}
	ret = make(Histogram[T], binCount)

	// create edges.  make sure last is equal to max
	if beforeCalc == nil {
		var spacing = (max - min) / T(binCount)
		for i := 0; i < len(ret); i++ {
			ret[i].Low = min + T(i)*spacing
		}
	} else {
		var spacing = (beforeCalc(max) - beforeCalc(min)) / float64(binCount)
		for i := 0; i < len(ret); i++ {
			ret[i].Low = afterCalc(beforeCalc(min) + float64(i)*spacing)
		}
	}
	for i := 0; i < len(ret)-1; i++ {
		ret[i].High = ret[i+1].Low
	}
	ret[binCount-1].High = max

	// calculate bins
	for _, v := range inSlice {
		var bucketIndex int
		if v > ret[0].Low {
			bucketIndex = sort.Search(len(ret), func(i int) bool { return ret[i].Low > v }) - 1
		}
		ret[bucketIndex].add(v)
	}

	return
}

// PrintOptions specifies options for the Print function
type PrintOptions[T constraints.Float | constraints.Integer] struct {
	Prefix    string
	Width     int
	Symbol    rune
	PadSymbol rune
	Format    func(T) string
	// SumSymbol is used, if not zero, for plotting the sum of a bin
	SumSymbol rune
	// CombinedSymbol is used, if not zero, for plotting the sum and count
	// of a bin, when they appear at the same position in the plot
	CombinedSymbol rune
}

// GetBytesOptions returns a PrintOptions for printout out byte sizes
func GetBytesOptions[T constraints.Integer](width int, printSum bool) PrintOptions[T] {
	var ret PrintOptions[T] = PrintOptions[T]{
		Width:     width,
		Symbol:    '|',
		PadSymbol: '-',
		Prefix:    "  ",
		Format:    _BytesPad[T],
	}
	if printSum {
		ret.SumSymbol = 'S'
		ret.CombinedSymbol = '$'
	}
	return ret
}

// func Print[T constraints.Float | constraints.Integer](w io.Writer, hist Histogram[T], opt PrintOptions) {}

// Print prints a histogram to a writer
func Print[T constraints.Float | constraints.Integer](w io.Writer, h Histogram[T], opt PrintOptions[T]) {
	if opt.Format == nil {
		opt.Format = func(i T) string { return fmt.Sprintf("%v", i) }
	}
	if len(h) == 0 {
		return
	}
	// normalize counts, sums to width
	counts := make([]int, len(h))
	sums := make([]T, len(counts))
	for i := 0; i < len(counts); i++ {
		counts[i] = h[i].Count
		sums[i] = h[i].Sum
	}
	countsNorm := normToWidth(counts, opt.Width)
	sumsNorm := normToWidth(sums, opt.Width)
	maxCountStr := strconv.Itoa(maxSlice(counts))
	// max line
	fmt.Fprintf(w, "%s  %s\n", opt.Prefix, opt.Format(h[len(h)-1].High))
	for i := len(h) - 1; i >= 0; i-- {
		b := bar(opt.Width, countsNorm[i], sumsNorm[i], opt.PadSymbol, opt.Symbol, opt.SumSymbol, opt.CombinedSymbol)
		var sumText string = strconv.Itoa(counts[i])
		for len(sumText) < len(maxCountStr) {
			sumText = " " + sumText
		}
		if opt.SumSymbol != 0 {
			sumText += "; " + opt.Format(sums[i])
		}
		fmt.Fprintf(w, "%s> %s: %s (%s)\n", opt.Prefix, opt.Format(h[i].Low), b, sumText)
	}
}

// normToWidth returns an array of normed values ranging from [0,width)
func normToWidth[T constraints.Float | constraints.Integer](s []T, width int) []int {
	max := float64(maxSlice(s))
	normed := make([]int, 0, len(s))
	for _, v := range s {
		normed = append(normed, int(math.Round(float64(v)/max*(float64(width)-1))))
	}
	return normed
}

func bar(width, val1, val2 int, val1pad, char1, char2, combinedChar rune) (ret string) {
	if val1pad == 0 {
		val1pad = ' '
	}
	if val1 > 0 {
		ret = strings.Repeat(string(val1pad), val1)
	}
	ret += string(char1)
	for len(ret) < width {
		ret += " "
	}

	if char2 != 0 {
		if val1 == val2 && combinedChar != 0 {
			char2 = combinedChar
		}
		ret = ret[:val2] + string(char2) + ret[val2+1:]
	}

	return
}
