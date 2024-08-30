package game

import (
	"RogueUI/foundation"
	"bytes"
	"encoding/gob"
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

func (r *ReadableObject) GobEncode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	if err := r.BaseObject.gobEncode(enc); err != nil {
		return nil, err
	}

	if err := enc.Encode(r.text); err != nil {
		return nil, err
	}

	if err := enc.Encode(r.textFile); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (r *ReadableObject) GobDecode(data []byte) error {
	dec := gob.NewDecoder(bytes.NewReader(data))

	r.BaseObject = &BaseObject{}

	if err := r.BaseObject.gobDecode(dec); err != nil {
		return err
	}
	if err := dec.Decode(&r.text); err != nil {
		return err
	}
	if err := dec.Decode(&r.textFile); err != nil {
		return err
	}

	return nil
}
func (g *GameState) NewReadable(rec recfile.Record) *ReadableObject {
	sign := &ReadableObject{
		BaseObject: NewObject(foundation.ObjectReadable, g.iconForObject),
	}

	sign.SetWalkable(false)
	sign.SetHidden(false)
	sign.SetTransparent(true)

	var customIcon textiles.TextIcon
	for _, field := range rec {
		switch strings.ToLower(field.Name) {
		case "name":
			customIcon = g.iconForObject(field.Value)
		case "icon":
			customIcon.Char = field.AsRune()
		case "fg":
			customIcon.Fg = field.AsRGB(",")
		case "bg":
			customIcon.Bg = field.AsRGB(",")
		case "description":
			sign.displayName = field.Value
		case "position":
			spawnPos, _ := geometry.NewPointFromEncodedString(field.Value)
			sign.SetPosition(spawnPos)
		case "text":
			sign.text = field.Value
		case "textfile":
			sign.textFile = field.Value
		}
	}

	sign.customIcon = customIcon
	sign.useCustomIcon = true
	sign.internalName = "readable"
	sign.InitWithGameState(g)
	return sign
}

func (r *ReadableObject) InitWithGameState(g *GameState) {
	r.iconForObject = g.iconForObject
	r.isPlayer = func(actor *Actor) bool { return actor == g.Player }
	r.showText = func(shown string) {
		g.ui.OpenTextWindow(g.fillTemplatedText(shown))
	}
	r.showTextFile = func(file string) {
		shown := fxtools.ReadFile(path.Join(g.config.DataRootDir, "text", file+".txt"))
		g.ui.OpenTextWindow(g.fillTemplatedText(shown))
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
		{Name: "category", Value: r.category.String()},
		{Name: "position", Value: r.position.Encode()},
		{Name: "description", Value: r.displayName},
		{Name: "icon", Value: string(r.customIcon.Char)},
		{Name: "fg", Value: recfile.RGBStr(r.customIcon.Fg)},
		{Name: "bg", Value: recfile.RGBStr(r.customIcon.Bg)},
	}
	if r.text != "" {
		rec = append(rec, recfile.Field{Name: "text", Value: r.text})
	}
	if r.textFile != "" {
		rec = append(rec, recfile.Field{Name: "textfile", Value: r.textFile})
	}
	return rec
}
