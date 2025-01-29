package ui_graphics

import (
    "fmt"
    "github.com/memmaker/go/geometry"
    "github.com/memmaker/go/recfile"
    "io/fs"
    "os"
)

type ItemNamingScheme string

const (
    ItemNamingSchemeShort        ItemNamingScheme = "Short"
    ItemNamingSchemeEnchantments ItemNamingScheme = "Enchantments"
    ItemNamingSchemeBandStyle    ItemNamingScheme = "BandStyle"
)

type FontInfo struct {
    TTFFile string
    TTFSize float64
}

type StartInfo struct {
    FontInfos         FontInfo
    ScreenSize        geometry.Point
    TileScale         float64
    Title             string
    LDTKMapFile       string
    UserConfiguration Configuration
    FileSystem        fs.ReadDirFS
}

type Configuration struct {
    AudioEnabled         bool
    ActorBobbingEnabled  bool
    WindowSize           geometry.Point
    DefaultFont          string
    DefaultFontSize      float64
    AutoPickup           bool
    CardinalMovementOnly bool
    KeepPlayerCentered   bool

    ItemNamesInShop ItemNamingScheme

    DrawTilesBehindActors  bool
    DrawTilesBehindItems   bool
    DrawTilesBehindObjects bool
    ServeStats             bool
    ServeInventory         bool
    ServeLog               bool
    ServeEquipment         bool
    FlipActorsHorizontally bool
    TargetingIconIndex     int32
}

func (c Configuration) ToRecord() recfile.Record {
    return recfile.Record{
        {
            Name:  "AudioEnabled",
            Value: recfile.BoolStr(c.AudioEnabled),
        },
        {
            Name:  "ActorBobbingEnabled",
            Value: recfile.BoolStr(c.ActorBobbingEnabled),
        },
        {
            Name:  "DrawTilesBehindActors",
            Value: recfile.BoolStr(c.DrawTilesBehindActors),
        },
        {
            Name:  "DrawTilesBehindItems",
            Value: recfile.BoolStr(c.DrawTilesBehindItems),
        },
        {
            Name:  "DrawTilesBehindObjects",
            Value: recfile.BoolStr(c.DrawTilesBehindObjects),
        },
        {
            Name:  "DefaultFont",
            Value: c.DefaultFont,
        },
        {
            Name:  "DefaultFontSize",
            Value: recfile.FloatStr(c.DefaultFontSize),
        },
        {
            Name:  "AutoPickup",
            Value: recfile.BoolStr(c.AutoPickup),
        },
        {
            Name:  "WindowWidth",
            Value: recfile.IntStr(c.WindowSize.X),
        },
        {
            Name:  "WindowHeight",
            Value: recfile.IntStr(c.WindowSize.Y),
        },
        {
            Name:  "CardinalMovementOnly",
            Value: recfile.BoolStr(c.CardinalMovementOnly),
        },
        {
            Name:  "KeepPlayerCentered",
            Value: recfile.BoolStr(c.KeepPlayerCentered),
        },
        {
            Name:  "ServeStats",
            Value: recfile.BoolStr(c.ServeStats),
        },
        {
            Name:  "ServeInventory",
            Value: recfile.BoolStr(c.ServeInventory),
        },
        {
            Name:  "ServeEquipment",
            Value: recfile.BoolStr(c.ServeEquipment),
        },
        {
            Name:  "ServeLog",
            Value: recfile.BoolStr(c.ServeLog),
        },
        {
            Name:  "ItemNamesInShop",
            Value: string(c.ItemNamesInShop),
        },
        {
            Name:  "FlipActorsHorizontally",
            Value: recfile.BoolStr(c.FlipActorsHorizontally),
        },
        {
            Name:  "TargetingIconIndex",
            Value: recfile.Int32Str(c.TargetingIconIndex),
        },
    }
}

func (c Configuration) SaveToRecFile(recFilename string) error {
    file, err := os.Create(recFilename)
    if err != nil {
        panic(err)
    }
    defer file.Close()
    return recfile.Write(file, []recfile.Record{c.ToRecord()})
}

func NewConfigurationFromRecFile(recFilename string) Configuration {
    println(fmt.Sprintf("Loading configuration from %s", recFilename))
    file, err := os.Open(recFilename)
    configExists := false
    var configMap recfile.DataMap

    if err == nil {
        records, _ := recfile.Read(file)
        if len(records) > 0 {
            configMap = records[0].ToMap(",")
            configExists = true
        }
    }

    if configMap == nil {
        configMap = make(recfile.DataMap)
    }

    configuration := Configuration{
        AudioEnabled:           configMap.GetBoolOrFalse("AudioEnabled"),
        ActorBobbingEnabled:    configMap.GetBoolOrTrue("ActorBobbingEnabled"),
        CardinalMovementOnly:   configMap.GetBoolOrFalse("CardinalMovementOnly"),
        DefaultFont:            configMap.GetStringOrDefault("DefaultFont", "assets/fonts/Ac437_EagleSpCGA_Alt3.ttf"),
        DefaultFontSize:        configMap.GetFloatOrDefault("DefaultFontSize", 16.0),
        AutoPickup:             configMap.GetBoolOrTrue("AutoPickup"),
        KeepPlayerCentered:     configMap.GetBoolOrTrue("KeepPlayerCentered"),
        DrawTilesBehindActors:  configMap.GetBoolOrFalse("DrawTilesBehindActors"),
        DrawTilesBehindItems:   configMap.GetBoolOrFalse("DrawTilesBehindItems"),
        DrawTilesBehindObjects: configMap.GetBoolOrFalse("DrawTilesBehindObjects"),
        ServeStats:             configMap.GetBoolOrFalse("ServeStats"),
        ServeInventory:         configMap.GetBoolOrFalse("ServeInventory"),
        ServeEquipment:         configMap.GetBoolOrFalse("ServeEquipment"),
        ServeLog:               configMap.GetBoolOrFalse("ServeLog"),
        FlipActorsHorizontally: configMap.GetBoolOrTrue("FlipActorsHorizontally"),
        ItemNamesInShop:        ItemNamingScheme(configMap.GetStringOrDefault("ItemNamesInShop", string(ItemNamingSchemeEnchantments))),
        WindowSize: geometry.Point{ //1280, Y: 800
            X: configMap.GetIntOrDefault("WindowWidth", 1280),
            Y: configMap.GetIntOrDefault("WindowHeight", 800),
        },
        TargetingIconIndex: int32(configMap.GetIntOrDefault("TargetingIconIndex", 0)),
    }
    if !configExists {
        configuration.SaveToRecFile(recFilename)
    }
    return configuration
}
