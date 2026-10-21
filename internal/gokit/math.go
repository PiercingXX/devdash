package gokit

import (
	"math"
	"strings"
)

func Round(n float64, precision int) (r float64) {
	pow := math.Pow(10, float64(precision))
	num := n * pow
	_, div := math.Modf(num)

	if n >= 0 && div >= 0.5 {
		r = math.Ceil(num)
	} else if n < 0 && div > -0.5 {
		r = math.Ceil(num)
	} else {
		r = math.Floor(num)
	}

	return r / pow
}

// ConvertBinUnit converts a value from a binary unit to another: b, kb, mb, gb, tb, pb or eb (powers of 1024).
// An unknown unit is considered as bytes.
func ConvertBinUnit(val float64, base, to string) float64 {
	convert := map[string]int{
		"b":  0,
		"kb": 1,
		"mb": 2,
		"gb": 3,
		"tb": 4,
		"pb": 5,
		"eb": 6,
	}

	r := convert[strings.ToLower(to)] - convert[strings.ToLower(base)]

	return Round(val/math.Pow(1024, float64(r)), 2)
}

func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
