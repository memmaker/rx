package d100

import "fmt"

type Percentage float32

func (p Percentage) String() string {
	return fmt.Sprintf("%d%%", int(p))
}

func (p Percentage) Normalized() float64 {
	return float64(p) / 100.0
}
