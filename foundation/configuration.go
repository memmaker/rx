package foundation

import (
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"os"
	"time"
)

type Configuration struct {
	AnimationDelay             time.Duration
	AnimationsEnabled          bool
	AnimateMovement            bool
	AnimateProjectiles         bool
	AnimateDamage              bool
	AnimateEffects             bool
	MapWidth                   int
	MapHeight                  int
	DiagonalMovementEnabled    bool
	AutoPickup                 bool
	PlayerName                 string
	WallSlide                  bool
	DataRootDir                string
	SaveGameDir                string
	DefaultToAdvancedTargeting bool
	PlayerChar                 rune
	PlayerColor                string
	KeyMap                     string

	DialogueShortcutsAreNumbers bool

	AudioEnabled          bool
	MusicEnabled          bool
	SoundEffectsEnabled   bool
	MainFontName          string
	FallbackFontName      string
	ForcedFallbackRunes   string
	SimulateAllLoadedMaps bool
	TileScale             float64
	TileWidth             int
	TileHeight            int
	WhiteTileIndex        int32
	TargetingTileIndex    int32
	KeepPlayerCentered    bool
	WindowSize            geometry.Point
}

func NewConfigurationFromFile(file string) *Configuration {
	configuration := NewDefaultConfiguration()
	if !fxtools.FileExists(file) {
		configuration.WriteToFile(file)
		return configuration
	}
	openFile := fxtools.MustOpen(file)
	data, _ := recfile.ReadAndClose(openFile)
	for _, field := range data[0] {
		switch field.Name {
		case "AnimationDelay":
			configuration.AnimationDelay = time.Duration(field.AsFloat()) * time.Millisecond
		case "AnimationsEnabled":
			configuration.AnimationsEnabled = field.AsBool()
		case "AnimateMovement":
			configuration.AnimateMovement = field.AsBool()
		case "AnimateProjectiles":
			configuration.AnimateProjectiles = field.AsBool()
		case "AnimateDamage":
			configuration.AnimateDamage = field.AsBool()
		case "AnimateEffects":
			configuration.AnimateEffects = field.AsBool()
		case "AudioEnabled":
			configuration.AudioEnabled = field.AsBool()
		case "MusicEnabled":
			configuration.MusicEnabled = field.AsBool()
		case "SoundEffectsEnabled":
			configuration.SoundEffectsEnabled = field.AsBool()
		case "MapWidth":
			configuration.MapWidth = field.AsInt()
		case "MapHeight":
			configuration.MapHeight = field.AsInt()
		case "DiagonalMovementEnabled":
			configuration.DiagonalMovementEnabled = field.AsBool()
		case "AutoPickup":
			configuration.AutoPickup = field.AsBool()
		case "WallSlide":
			configuration.WallSlide = field.AsBool()
		case "DataRootDir":
			configuration.DataRootDir = field.Value
		case "SaveGameDir":
			configuration.SaveGameDir = field.Value
		case "DefaultToAdvancedTargeting":
			configuration.DefaultToAdvancedTargeting = field.AsBool()
		case "PlayerName":
			configuration.PlayerName = field.Value
		case "PlayerChar":
			configuration.PlayerChar = field.AsRune()
		case "PlayerColor":
			configuration.PlayerColor = field.Value
		case "KeyMap":
			configuration.KeyMap = field.Value
		case "DialogueShortcutsAreNumbers":
			configuration.DialogueShortcutsAreNumbers = field.AsBool()
		case "MainFontName":
			configuration.MainFontName = field.Value
		case "FallbackFontName":
			configuration.FallbackFontName = field.Value
		case "ForcedFallbackRunes":
			configuration.ForcedFallbackRunes = field.Value
		case "ForcedFallbackCodepoint":
			configuration.ForcedFallbackRunes += string(rune(field.AsInt()))
		case "SimulateAllLoadedMaps":
			configuration.SimulateAllLoadedMaps = field.AsBool()
		case "TileScale":
			configuration.TileScale = field.AsFloat()
		case "TileWidth":
			configuration.TileWidth = field.AsInt()
		case "TileHeight":
			configuration.TileHeight = field.AsInt()
		case "WhiteTileIndex":
			configuration.WhiteTileIndex = field.AsInt32()
		case "TargetingTileIndex":
			configuration.TargetingTileIndex = field.AsInt32()
		case "KeepPlayerCentered":
			configuration.KeepPlayerCentered = field.AsBool()
		case "WindowSize":
			configuration.WindowSize = geometry.MustDecodePoint(field.Value)
		}
	}
	return configuration
}
func NewDefaultConfiguration() *Configuration {
	return &Configuration{
		AnimationDelay:              55 * time.Millisecond,
		AnimationsEnabled:           true,
		AnimateMovement:             false,
		AnimateProjectiles:          true,
		AnimateDamage:               true,
		AnimateEffects:              true,
		MapWidth:                    80,
		MapHeight:                   23,
		DiagonalMovementEnabled:     true,
		AutoPickup:                  true,
		PlayerName:                  "Rogue",
		WallSlide:                   true,
		DataRootDir:                 "data_atom",
		SaveGameDir:                 "save",
		DefaultToAdvancedTargeting:  true,
		PlayerChar:                  '@',
		PlayerColor:                 "white",
		KeyMap:                      "numpad",
		DialogueShortcutsAreNumbers: false,
		AudioEnabled:                true,
		MusicEnabled:                true,
		SoundEffectsEnabled:         true,
		MainFontName:                "Monofonto-Regular",
		FallbackFontName:            "MesloLGS NF Regular",
		SimulateAllLoadedMaps:       true,
		TileScale:                   2,
		TileWidth:                   8,
		TileHeight:                  8,
		WhiteTileIndex:              0,
		TargetingTileIndex:          1,
		KeepPlayerCentered:          true,
		WindowSize:                  geometry.Point{X: 800, Y: 600},
	}
}

func (c *Configuration) GetMinTerminalSize() (int, int) {
	return c.MapWidth, c.MapHeight + 2
}

func (c *Configuration) WriteToFile(filename string) {
	record := recfile.Record{
		recfile.Field{Name: "AnimationDelay", Value: recfile.Int64Str(c.AnimationDelay.Milliseconds())},
		recfile.Field{Name: "AnimationsEnabled", Value: recfile.BoolStr(c.AnimationsEnabled)},
		recfile.Field{Name: "AnimateMovement", Value: recfile.BoolStr(c.AnimateMovement)},
		recfile.Field{Name: "AnimateProjectiles", Value: recfile.BoolStr(c.AnimateProjectiles)},
		recfile.Field{Name: "AnimateDamage", Value: recfile.BoolStr(c.AnimateDamage)},
		recfile.Field{Name: "AnimateEffects", Value: recfile.BoolStr(c.AnimateEffects)},
		recfile.Field{Name: "AudioEnabled", Value: recfile.BoolStr(c.AudioEnabled)},
		recfile.Field{Name: "MusicEnabled", Value: recfile.BoolStr(c.MusicEnabled)},
		recfile.Field{Name: "MapWidth", Value: recfile.IntStr(c.MapWidth)},
		recfile.Field{Name: "MapHeight", Value: recfile.IntStr(c.MapHeight)},
		recfile.Field{Name: "DiagonalMovementEnabled", Value: recfile.BoolStr(c.DiagonalMovementEnabled)},
		recfile.Field{Name: "AutoPickup", Value: recfile.BoolStr(c.AutoPickup)},
		recfile.Field{Name: "WallSlide", Value: recfile.BoolStr(c.WallSlide)},
		recfile.Field{Name: "PlayerName", Value: c.PlayerName},
		recfile.Field{Name: "PlayerChar", Value: string(c.PlayerChar)},
		recfile.Field{Name: "PlayerColor", Value: c.PlayerColor},
		recfile.Field{Name: "DataRootDir", Value: c.DataRootDir},
		recfile.Field{Name: "SaveGameDir", Value: c.SaveGameDir},
		recfile.Field{Name: "DefaultToAdvancedTargeting", Value: recfile.BoolStr(c.DefaultToAdvancedTargeting)},
		recfile.Field{Name: "KeyMap", Value: c.KeyMap},
		recfile.Field{Name: "DialogueShortcutsAreNumbers", Value: recfile.BoolStr(c.DialogueShortcutsAreNumbers)},
		recfile.Field{Name: "MainFontName", Value: c.MainFontName},
		recfile.Field{Name: "FallbackFontName", Value: c.FallbackFontName},
		recfile.Field{Name: "ForcedFallbackRunes", Value: c.ForcedFallbackRunes},
		recfile.Field{Name: "SimulateAllLoadedMaps", Value: recfile.BoolStr(c.SimulateAllLoadedMaps)},
		recfile.Field{Name: "TileScale", Value: recfile.FloatStr(c.TileScale)},
		recfile.Field{Name: "TileWidth", Value: recfile.IntStr(c.TileWidth)},
		recfile.Field{Name: "TileHeight", Value: recfile.IntStr(c.TileHeight)},
		recfile.Field{Name: "WhiteTileIndex", Value: recfile.Int32Str(c.WhiteTileIndex)},
		recfile.Field{Name: "TargetingTileIndex", Value: recfile.Int32Str(c.TargetingTileIndex)},
		recfile.Field{Name: "KeepPlayerCentered", Value: recfile.BoolStr(c.KeepPlayerCentered)},
		recfile.Field{Name: "WindowSize", Value: c.WindowSize.Encode()},
	}
	file, _ := os.Create(filename)
	defer file.Close()
	recfile.Write(file, []recfile.Record{record})
}

func (c *Configuration) TileSize() geometry.Point {
	return geometry.Point{X: c.TileWidth, Y: c.TileHeight}
}
