package str

import (
	"math"
	"strconv"
	"strings"
)

// ChunkString chops string into as many equal size parts as possible,
// starting from left for positive (n), or right for negative (n).
//
// e.g. ChunkString("1234", 3)  -> ["123", "4"]
// e.g. Chunkstring("1234", -3) -> ["1", "234"]
func ChunkString(buf string, limit int) []string {
	if limit == 0 {
		return []string{buf}
	}

	var m, b, e int // mod, begin, end
	ll := len(buf)

	if limit < 0 {
		limit = 0 - limit
		m = ll % limit
	}

	ceiling := (ll + limit - 1) / limit
	parts := make([]string, ceiling)

	for i := range parts {
		if m == 0 { // every iteration EXCEPT i=0, limit < 0
			b, e = e, e+limit
			if e > ll {
				e = ll
			}
		} else { // first time thru for limit < 0
			b, e, m = 0, m, 0
		}

		parts[i] = buf[b:e]
	}

	return parts
}

// Comma creates a human readable integer by adding commas
// for thousands separators.
//
// e.g. Comma(1234567) -> 1,234,567
func Comma(n int64) string {
	// only special case
	if n == math.MinInt64 {
		return "-9,223,372,036,854,775,808"
	}

	sign := ""
	if n < 0 {
		sign = "-"
		n = 0 - n
	}

	parts := ChunkString(strconv.FormatInt(n, 10), -3)
	return sign + strings.Join(parts, ",")
}
