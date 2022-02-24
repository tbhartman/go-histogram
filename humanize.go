package histogram

import (
	"constraints"
	"fmt"
	"math"
	"strings"
)

var sizeStr = []string{
	"B",
	"KiB",
	"MiB",
	"GiB",
	"TiB",
	"PiB",
	"EiB",
	// beyond this is a uint64 overflow
}

func _BytesPad[T constraints.Integer](i T) string {
	s := _Bytes(i)
	s = strings.TrimSuffix(s, " B")
	if !strings.HasSuffix(s, "B") {
		s += "     B"
	}
	for len(s) < 10 {
		s = " " + s
	}
	return s
}

func _Bytes[T constraints.Integer](i T) string {
	if i == 0 {
		return "0 B"
	}
	pow := math.Log(float64(i)) / math.Log(1024)
	floor := int(math.Floor(pow))
	if floor >= len(sizeStr) {
		floor = len(sizeStr) - 1
	}
	if floor == 0 {
		return fmt.Sprintf("%d B", i)
	}
	val := float64(i) / math.Pow(1024, float64(floor))
	return fmt.Sprintf("%.1f %s", val, sizeStr[floor])
}
