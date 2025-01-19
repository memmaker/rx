package game

import (
	"contractor/foundation"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/geometry"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"path"
	"strings"
)

type ReadableObject struct {
	*BaseObject
	isPlayer     func(*Actor) bool
	showTextFile func(string)
	showText     func(string)
	text         string
	textFile     string
}

func (g *GameState) NewReadable(rec recfile.Record, resolver func(objType string) textiles.TextIcon) *ReadableObject {
	sign := &ReadableObject{
		BaseObject: NewObject(foundation.ObjectReadable, resolver),
	}

	sign.SetWalkable(false)
	sign.SetHidden(false)
	sign.SetTransparent(true)

	var customIcon textiles.TextIcon
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			sign.InternalName = field.Value
		case "iconoverride":
			customIcon = sign.iconForObject(field.Value)
		case "icon":
			customIcon.Char = field.AsRune()
		case "fg":
			customIcon.Fg = field.AsRGB(",")
		case "bg":
			customIcon.Bg = field.AsRGB(",")
		case "description":
			sign.DisplayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			sign.SetPosition(spawnPos)
		case "text":
			sign.text = field.Value
		case "textfile":
			sign.textFile = field.Value
		}
	}

	sign.CustomIcon = customIcon
	sign.UseCustomIcon = true
	sign.InternalName = "readable"
	sign.InitWithGameState(g)
	return sign
}

func (r *ReadableObject) InitWithGameState(g *GameState) {
	r.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	r.showText = func(shown string) {
		g.ui.OpenTextWindow(g.FillTemplatedText(shown))
	}
	r.showTextFile = func(file string) {
		shown := fxtools.ReadFile(path.Join(g.config.DataRootDir, "text", file+".txt"))
		g.ui.OpenTextWindow(g.FillTemplatedText(shown))
	}
}
func (r *ReadableObject) AppendContextActions(actions []foundation.MenuItem, g *GameState) []foundation.MenuItem {
	return append(actions, foundation.MenuItem{
		Name:       "Read",
		Action:     r.showTextToPlayer,
		CloseMenus: true,
	})
}

func (r *ReadableObject) OnBump(actor *Actor) {
	if r.isPlayer(actor) {
		r.showTextToPlayer()
	}
}

func (r *ReadableObject) showTextToPlayer() {
	if r.textFile != "" {
		r.showTextFile(r.textFile)
	} else {
		r.showText(r.text)
	}
}

func (r *ReadableObject) ToRecord() recfile.Record {
	rec := recfile.Record{
		{Name: "category", Value: r.Category.String()},
		{Name: "position", Value: r.RawPosition.Encode()},
		{Name: "description", Value: r.DisplayName},
		{Name: "icon", Value: string(r.CustomIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(r.CustomIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(r.CustomIcon.Bg)},
	}
	if r.text != "" {
		rec = append(rec, recfile.Field{Name: "text", Value: r.text})
	}
	if r.textFile != "" {
		rec = append(rec, recfile.Field{Name: "textfile", Value: r.textFile})
	}
	return rec
}
