package histogram

import (
	"testing"
)

func TestHumanize(t *testing.T) {
	for _, v := range []struct {
		Value  uint64
		String string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{900, "900 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{600000, "585.9 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
		{1024 * 1024 * 1024 * 1024, "1.0 TiB"},
		{1024 * 1024 * 1024 * 1024 * 1024, "1.0 PiB"},
		{1024 * 1024 * 1024 * 1024 * 1024 * 1024, "1.0 EiB"},
	} {
		res := _Bytes(v.Value)
		if res != v.String {
			t.Errorf("%d formatted to '%s'; expected '%s'", v.Value, res, v.String)
		}
	}
}
