package util

func Clamp(min, max, value float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
