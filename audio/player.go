package audio

import (
	"fmt"
	"github.com/gopxl/beep"
	"github.com/gopxl/beep/speaker"
	"github.com/gopxl/beep/vorbis"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

type Player struct {
	loadedCues    map[string][]*beep.Buffer
	currentStream beep.StreamSeekCloser
	cuesLoaded    bool
}

func NewPlayer() *Player {
	p := &Player{}
	p.UnloadAllCues()
	p.init()
	return p
}
func (p *Player) StopAll() {
	speaker.Clear()
}
func (p *Player) Stream(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	streamer, _, err := vorbis.Decode(f)
	if err != nil {
		return err
	}
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		streamer.Close()
	})))
	p.currentStream = streamer
	return nil
}
func (p *Player) UnloadAllCues() {
	if p.currentStream != nil {
		p.currentStream.Close()
	}
	p.loadedCues = make(map[string][]*beep.Buffer)
}
func (p *Player) LoadEnemyCuesFromDir(enemySfxDir string) {
	entries, _ := os.ReadDir(enemySfxDir)
	for _, entry := range entries {
		if entry.IsDir() {
			enemyName := entry.Name()
			p.LoadCuesFromDir(path.Join(enemySfxDir, enemyName), "enemies")
		}
	}
}

func (p *Player) LoadCuesFromDir(dirName string, cuePrefix string) {
	cueBase := path.Base(dirName)
	if cuePrefix != "" {
		cueBase = fmt.Sprintf("%s/%s", cuePrefix, cueBase)
	}
	entries, _ := os.ReadDir(dirName)
	for _, entry := range entries {
		if entry.IsDir() {
			cueName := fmt.Sprintf("%s/%s", cueBase, entry.Name())
			p.loadAllCuesInDir(path.Join(dirName, entry.Name()), cueName)
		} else {
			extension := path.Ext(entry.Name())
			if strings.ToLower(extension) != ".ogg" {
				continue
			}
			withoutExt := entry.Name()[:len(entry.Name())-len(extension)]
			cueName := fmt.Sprintf("%s/%s", cueBase, withoutExt)
			p.loadCue(path.Join(dirName, entry.Name()), cueName)
		}
	}
}
func (p *Player) loadCue(fileName string, cueName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}

	streamer, format, err := vorbis.Decode(f)
	if err != nil {
		return err
	}

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	closeErr := streamer.Close()
	if closeErr != nil {
		return closeErr
	}

	if _, exists := p.loadedCues[cueName]; !exists {
		p.loadedCues[cueName] = make([]*beep.Buffer, 0)
	}
	p.loadedCues[cueName] = append(p.loadedCues[cueName], buffer)
	return nil
}

func (p *Player) PlayCue(name string) {
	if !p.cuesLoaded {
		return
	}
	buffers, bufExists := p.loadedCues[name]
	if !bufExists || len(buffers) == 0 {
		return
	}

	randomIndex := 0
	if len(buffers) > 1 {
		randomIndex = rand.Intn(len(buffers))
	}

	buffer := buffers[randomIndex]
	//format := buffer.Format()

	sfx := buffer.Streamer(0, buffer.Len())
	speaker.Play(sfx)
}

func (p *Player) init() {
	sampleRate := beep.SampleRate(22050)
	initErr := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	if initErr != nil {
		fmt.Fprintf(os.Stderr, "Error initializing speaker: %v", initErr)
		return
	}
}

func (p *Player) loadAllCuesInDir(dirName string, cueName string) {
	entries, _ := os.ReadDir(dirName)
	for _, entry := range entries {
		if strings.ToLower(path.Ext(entry.Name())) != ".ogg" {
			continue
		}
		p.loadCue(path.Join(dirName, entry.Name()), cueName)
	}
}

func (p *Player) SoundsLoaded() {
	p.cuesLoaded = true
}
