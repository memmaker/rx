package game

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
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
	runningScripts []*ScriptInstance
}

func NewScriptRunner() *ScriptRunner {
	return &ScriptRunner{
		runningScripts: make([]*ScriptInstance, 0),
	}
}

func (s *ScriptRunner) RunScriptByName(dataDir string, scriptName string, condFuncs map[string]govaluate.ExpressionFunction) {
	script := LoadScript(dataDir, scriptName, condFuncs)
	s.RunScript(script)
}

func (s *ScriptRunner) RunScript(script ActionScript) {
	runningScript := &ScriptInstance{
		script:       script,
		currentFrame: 0,
	}
	s.runningScripts = append(s.runningScripts, runningScript)
}

func (s *ScriptRunner) CheckAndRunFrames() {
	s.runningScripts = fxtools.FilterSlice(s.runningScripts, func(instance *ScriptInstance) bool {
		return !instance.IsDone()
	})

	for _, instance := range s.runningScripts {
		if endFrame, isValid := instance.HasReachedOutcome(); isValid {
			instance.Run(endFrame)
			instance.SetDone()
		} else if instance.CanRunCurrentFrame() {
			instance.RunCurrentFrame()
		}
	}
}

func (s *ScriptRunner) StopScript(name string) {
	scriptIndex := -1
	for index, instance := range s.runningScripts {
		if instance.script.Name == name {
			scriptIndex = index
			break
		}
	}

	if scriptIndex != -1 {
		scripToStop := s.runningScripts[scriptIndex]
		scripToStop.RunCancelFrame()
		s.runningScripts = append(s.runningScripts[:scriptIndex], s.runningScripts[scriptIndex+1:]...)
	}
}
func (s *ScriptRunner) String() string {
	if len(s.runningScripts) == 0 {
		return "No running scripts"
	}
	out := make([]string, len(s.runningScripts))
	for i, instance := range s.runningScripts {
		out[i] = fmt.Sprintf("%s @ %d: %s", instance.script.Name, instance.currentFrame, instance.GetCurrentFrame().String())
	}
	return "Running scripts:\n" + strings.Join(out, "\n")
}
