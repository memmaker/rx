package ui_graphics

import (
    "fmt"
    "github.com/hajimehoshi/ebiten/v2"
    "image/color"
    "math"
    "regexp"
)

func RemoveColorCodes(colorCodedText string) string {
    // color codes look like this [:red] [:blue] [:green] [:yellow] [:white] [:24,54,222]
    regexPattern := `\[(:[\w,]+)\]`
    rp, _ := regexp.Compile(regexPattern)

    return rp.ReplaceAllString(colorCodedText, "")
}

func GetShortCutForIndex(i int) rune {
    if i < 9 {
        return rune('1' + i)
    }
    // else start with a
    return rune('A' + i - 9)
}

func SecondsToTicks(seconds float64) uint64 {
    return uint64(secondsToTicksF(seconds))
}

func secondsToTicksF(seconds float64) float64 {
    return safeTPS() * seconds
}

func safeTPS() float64 {
    return max(60, ebiten.ActualTPS())
}

func Clamp(value float64, minVal float64, maxVal float64) float64 {
    return min(maxVal, max(minVal, value))
}

func RGBAToColorCode(color color.RGBA) string {
    return fmt.Sprintf("[:%d,%d,%d]", color.R, color.G, color.B)
}

func EaseInExpo(x float64) float64 {
    if x == 0 {
        return 0
    }
    return math.Pow(2, 10*x-10)
}
