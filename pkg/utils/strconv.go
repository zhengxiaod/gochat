package utils

import "strconv"

// Uint64ToStr uint64 -> str
func Uint64ToStr(num uint64) string {
	return strconv.FormatUint(num, 10)
}
