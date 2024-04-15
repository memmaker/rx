package foundation

import (
	"RogueUI/recfile"
	"RogueUI/util"
	"os"
	"path"
	"time"
)

type Configuration struct {
	AnimationDelay          time.Duration
	AnimationsEnabled       bool
	AnimateMovement         bool
	AnimateProjectiles      bool
	AnimateDamage           bool
	AnimateEffects          bool
	MapWidth                int
	MapHeight               int
	DiagonalMovementEnabled bool
	AutoPickup              bool
	PlayerName              string
	KeyMapFile              string
	Theme                   string
	WallSlide               bool
}

func NewConfigurationFromFile(file string) *Configuration {
	configuration := NewDefaultConfiguration()
	if !util.FileExists(file) {
		configuration.WriteToFile(file)
		return configuration
	}
	openFile := util.MustOpen(file)
	defer openFile.Close()
	data := recfile.Read(openFile)
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
		case "KeyMapFile":
			configuration.KeyMapFile = field.Value
		case "Theme":
			configuration.Theme = field.Value
		}
	}
	return configuration
}
func NewDefaultConfiguration() *Configuration {
	return &Configuration{
		MapWidth:                80,
		MapHeight:               23,
		DiagonalMovementEnabled: true,
		AnimationDelay:          75 * time.Millisecond,

		AnimationsEnabled:  true,
		AnimateMovement:    true,
		AnimateDamage:      true,
		AnimateEffects:     true,
		AnimateProjectiles: true,
		AutoPickup:         true,
		WallSlide:          true,
		PlayerName:         "Rogue",
		KeyMapFile:         path.Join("data", "keymaps", "default.txt"),
		Theme:              path.Join("data", "themes", "ascii.rec"),
	}
}

func (c *Configuration) GetMinTerminalSize() (int, int) {
	return c.MapWidth, c.MapHeight + 2
}

func (c *Configuration) WriteToFile(filename string) {
	records := []recfile.Record{
		{
			recfile.Field{Name: "AnimationDelay", Value: recfile.Int64Str(c.AnimationDelay.Milliseconds())},
			recfile.Field{Name: "AnimationsEnabled", Value: recfile.BoolStr(c.AnimationsEnabled)},
			recfile.Field{Name: "AnimateMovement", Value: recfile.BoolStr(c.AnimateMovement)},
			recfile.Field{Name: "AnimateProjectiles", Value: recfile.BoolStr(c.AnimateProjectiles)},
			recfile.Field{Name: "AnimateDamage", Value: recfile.BoolStr(c.AnimateDamage)},
			recfile.Field{Name: "AnimateEffects", Value: recfile.BoolStr(c.AnimateEffects)},
			recfile.Field{Name: "MapWidth", Value: recfile.IntStr(c.MapWidth)},
			recfile.Field{Name: "MapHeight", Value: recfile.IntStr(c.MapHeight)},
			recfile.Field{Name: "DiagonalMovementEnabled", Value: recfile.BoolStr(c.DiagonalMovementEnabled)},
			recfile.Field{Name: "AutoPickup", Value: recfile.BoolStr(c.AutoPickup)},
			recfile.Field{Name: "KeyMapFile", Value: c.KeyMapFile},
		},
	}
	file, _ := os.Create(filename)
	defer file.Close()
	recfile.Write(file, records)
}
