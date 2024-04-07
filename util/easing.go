package util

import "math"

/*

function easeInExpo(x: number): number {
return x === 0 ? 0 : Math.pow(2, 10 * x - 10);
}
*/

func EaseInExpo(x float64) float64 {
	if x == 0 {
		return 0
	}
	return math.Pow(2, 10*x-10)
}
