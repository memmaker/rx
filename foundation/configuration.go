package foundation

import (
	"github.com/memmaker/go/fxtools"
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
	DefaultToAdvancedTargeting bool
	PlayerChar                 rune
	PlayerColor                string
	KeyMap                     string

	DialogueShortcutsAreNumbers bool
	UseLockpickingMiniGame      bool
	UseLockpickingDX            bool
}

func NewConfigurationFromFile(file string) *Configuration {
	configuration := NewDefaultConfiguration()
	if !fxtools.FileExists(file) {
		configuration.WriteToFile(file)
		return configuration
	}
	openFile := fxtools.MustOpen(file)
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
		case "DataRootDir":
			configuration.DataRootDir = field.Value
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
		case "UseLockpickingMiniGame":
			configuration.UseLockpickingMiniGame = field.AsBool()
		case "UseLockpickingDX":
			configuration.UseLockpickingDX = field.AsBool()
		}
	}
	return configuration
}
func NewDefaultConfiguration() *Configuration {
	return &Configuration{
		MapWidth:                80,
		MapHeight:               23,
		DiagonalMovementEnabled: true,
		AnimationDelay:          55 * time.Millisecond,

		AnimationsEnabled:           true,
		AnimateMovement:             false,
		AnimateDamage:               true,
		AnimateEffects:              true,
		AnimateProjectiles:          true,
		AutoPickup:                  true,
		WallSlide:                   true,
		DataRootDir:                 "data_atom",
		DefaultToAdvancedTargeting:  true,
		PlayerName:                  "Rogue",
		PlayerChar:                  '@',
		PlayerColor:                 "white",
		KeyMap:                      "numpad",
		DialogueShortcutsAreNumbers: false,
		UseLockpickingMiniGame:      false,
		UseLockpickingDX:            false,
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
			recfile.Field{Name: "WallSlide", Value: recfile.BoolStr(c.WallSlide)},
			recfile.Field{Name: "PlayerName", Value: c.PlayerName},
			recfile.Field{Name: "PlayerChar", Value: string(c.PlayerChar)},
			recfile.Field{Name: "PlayerColor", Value: c.PlayerColor},
			recfile.Field{Name: "DataRootDir", Value: c.DataRootDir},
			recfile.Field{Name: "DefaultToAdvancedTargeting", Value: recfile.BoolStr(c.DefaultToAdvancedTargeting)},
			recfile.Field{Name: "KeyMap", Value: c.KeyMap},
			recfile.Field{Name: "DialogueShortcutsAreNumbers", Value: recfile.BoolStr(c.DialogueShortcutsAreNumbers)},
			recfile.Field{Name: "UseLockpickingMiniGame", Value: recfile.BoolStr(c.UseLockpickingMiniGame)},
			recfile.Field{Name: "UseLockpickingDX", Value: recfile.BoolStr(c.UseLockpickingDX)},
		},
	}
	file, _ := os.Create(filename)
	defer file.Close()
	recfile.Write(file, records)
}
