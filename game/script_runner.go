package game

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
	"path"
	"strings"
)

type ScriptInstance struct {
	script       ActionScript
	currentFrame int
	isDone       bool
}

func (i *ScriptInstance) GetCurrentFrame() ScriptFrame {
	return i.script.Frames[i.currentFrame]
}

func (i *ScriptInstance) CanRunCurrentFrame() bool {
	if i.currentFrame >= len(i.script.Frames) {
		return false
	}
	frame := i.GetCurrentFrame()

	return i.script.CanRunFrame(frame)
}

func (i *ScriptInstance) RunCurrentFrame() {
	if i.currentFrame >= len(i.script.Frames) {
		return
	}
	i.Run(i.GetCurrentFrame())
	i.currentFrame++
}

func (i *ScriptInstance) Run(frame ScriptFrame) {
	frame.ExecuteActions(i.script.Variables)
}

func (i *ScriptInstance) HasReachedOutcome() (ScriptFrame, bool) {
	for _, outcomeFrame := range i.script.Outcomes {
		condition := outcomeFrame.Condition(i.script.Variables)
		if condition {
			return outcomeFrame, true
		}
	}
	return UserScriptFrame{}, false
}

func (i *ScriptInstance) IsDone() bool {
	return i.isDone
}

func (i *ScriptInstance) SetDone() {
	i.isDone = true
}

func (i *ScriptInstance) RunCancelFrame() {
	if i.script.CancelFrame.IsEmpty() {
		return
	}
	i.Run(i.script.CancelFrame)
}

type ScriptRunner struct {
	runningScripts map[string][]*ScriptInstance
}

func NewScriptRunner() *ScriptRunner {
	return &ScriptRunner{
		runningScripts: make(map[string][]*ScriptInstance),
	}
}

func (s *ScriptRunner) RunScriptByName(mapDir string, mapName string, scriptName string, condFuncs map[string]govaluate.ExpressionFunction) {
	script := LoadScript(path.Join(mapDir, mapName), scriptName, condFuncs)
	s.RunScript(mapName, script)
}

func (s *ScriptRunner) RunScript(currentMapName string, script ActionScript) {
	runningScript := &ScriptInstance{
		script:       script,
		currentFrame: 0,
	}
	s.runningScripts[currentMapName] = append(s.runningScripts[currentMapName], runningScript)
}

func (s *ScriptRunner) CheckAndRunFrames(mapName string) {
	s.runningScripts[mapName] = fxtools.FilterSlice(s.runningScripts[mapName], func(instance *ScriptInstance) bool {
		return !instance.IsDone()
	})

	for _, instance := range s.runningScripts[mapName] {
		if endFrame, isValid := instance.HasReachedOutcome(); isValid {
			instance.Run(endFrame)
			instance.SetDone()
		} else if instance.CanRunCurrentFrame() {
			instance.RunCurrentFrame()
		}
	}
}

func (s *ScriptRunner) StopScript(mapName string, name string) {
	scriptIndex := -1
	for index, instance := range s.runningScripts[mapName] {
		if instance.script.Name == name {
			scriptIndex = index
			break
		}
	}

	if scriptIndex != -1 {
		scripToStop := s.runningScripts[mapName][scriptIndex]
		scripToStop.RunCancelFrame()
		s.runningScripts[mapName] = append(s.runningScripts[mapName][:scriptIndex], s.runningScripts[mapName][scriptIndex+1:]...)
	}
}
func (s *ScriptRunner) String() string {
	if len(s.runningScripts) == 0 {
		return "No running scripts"
	}
	out := make([]string, 0)
	for mapName, runningScripts := range s.runningScripts {
		out = append(out, fmt.Sprintf("Map: %s", mapName))
		for _, instance := range runningScripts {
			out = append(out, fmt.Sprintf(" %s @ %d: %s", instance.script.Name, instance.currentFrame, instance.GetCurrentFrame().String()))
		}
	}

	return "Running scripts:\n" + strings.Join(out, "\n")
}
