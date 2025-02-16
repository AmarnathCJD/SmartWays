package modules

import "strconv"

var (
	convertInt = func(s string) int {
		i, _ := strconv.Atoi(s)
		return i
	}

	convertString = func(i int) string {
		return strconv.Itoa(i)
	}

	convertFloat = func(s string) float64 {
		f, _ := strconv.ParseFloat(s, 64)
		return f
	}

	convertBool = func(s string) bool {
		b, _ := strconv.ParseBool(s)
		return b
	}
)
