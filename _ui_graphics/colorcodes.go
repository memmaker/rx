package ui_graphics

import (
    "image/color"
    "regexp"
    "strconv"
)

type ColoredTextPart struct {
    Text    string
    Color   color.RGBA
    XOffset float64
    Height  float64
}

func GetColorFromRGBString(lightColor string) color.RGBA {
    // looks like "234 12 124"
    regexpPattern := regexp.MustCompile(`([0-9]{1,3}) +([0-9]{1,3}) +([0-9]{1,3})`)
    matches := regexpPattern.FindStringSubmatch(lightColor)
    if matches == nil {
        return color.RGBA{}
    }
    r, _ := strconv.Atoi(matches[1])
    g, _ := strconv.Atoi(matches[2])
    b, _ := strconv.Atoi(matches[3])
    return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}
