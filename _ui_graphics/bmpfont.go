package ui_graphics

type BitmapFont struct {
    atlas   TextureAtlas
    fontMap map[rune]uint16
}

func (f BitmapFont) IsLoaded() bool {
    return f.atlas.imageData != nil && f.fontMap != nil && len(f.fontMap) > 0
}

func NewBitmapFont(atlas TextureAtlas, fontMap map[rune]uint16) BitmapFont {
    return BitmapFont{
        atlas:   atlas,
        fontMap: fontMap,
    }
}

type FontAtlasDescription struct {
    IndexOfCapitalA int
    IndexOfSmallA   *int
    IndexOfZero     *int
    IndexOfOne      *int
    Chains          []SpecialCharacterChain
}
type SpecialCharacterChain struct {
    StartIndex int
    Characters []rune
}

func NewFontIndexFromDescription(desc FontAtlasDescription) map[rune]uint16 {
    result := map[rune]uint16{}

    indexOfCapitalA := desc.IndexOfCapitalA

    indexOfZero := indexOfCapitalA + 26
    if desc.IndexOfZero != nil {
        indexOfZero = *desc.IndexOfZero
    }

    for i := 0; i < 26; i++ {
        result[rune(i+65)] = uint16(indexOfCapitalA + i)
    }

    if desc.IndexOfOne != nil {
        for i := 0; i < 9; i++ {
            // 1..9, 0
            result[rune(i+49)] = uint16(*desc.IndexOfOne + i)
        }
        result[rune(48)] = uint16(*desc.IndexOfOne + 9)
    } else {
        for i := 0; i < 10; i++ {
            result[rune(i+48)] = uint16(indexOfZero + i)
        }
    }

    if desc.IndexOfSmallA != nil {
        indexOfSmallA := *desc.IndexOfSmallA
        for i := 0; i < 26; i++ {
            result[rune(i+97)] = uint16(indexOfSmallA + i)
        }
    }

    for _, chain := range desc.Chains {
        indexOfSpecialChain := chain.StartIndex
        for i, special := range chain.Characters {
            result[special] = uint16(indexOfSpecialChain + i)
        }
    }

    return result
}
