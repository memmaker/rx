package game

import (
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/fxtools"
)

type ScriptInstance struct {
	script       ActionScript
	currentFrame int
	isDone       bool
}

func (i *ScriptInstance) IsDone() bool {
	return i.isDone || i.currentFrame >= len(i.script.Frames)
}
func (i *ScriptInstance) GetCurrentFrame() ScriptFrame {
	return i.script.Frames[i.currentFrame]
}

func (i *ScriptInstance) CanRunCurrentFrame() bool {
	frame := i.GetCurrentFrame()
	condition, err := frame.Condition.Evaluate(i.script.Variables)
	if err != nil {
		panic(err)
	}
	return condition.(bool)
}

func (i *ScriptInstance) RunCurrentFrame() {
	if i.IsDone() {
		return
	}
	i.Run(i.GetCurrentFrame())
	i.currentFrame++
}

func (i *ScriptInstance) Run(frame ScriptFrame) {
	for _, action := range frame.Actions {
		_, err := action.Evaluate(i.script.Variables)
		if err != nil {
			panic(err)
		}
	}
}

func (i *ScriptInstance) HasReachedOutcome() (ScriptFrame, bool) {
	for _, outcomeFrame := range i.script.Outcomes {
		condition, err := outcomeFrame.Condition.Evaluate(i.script.Variables)
		if err != nil {
			panic(err)
		}
		if condition.(bool) {
			return outcomeFrame, true
		}
	}
	return ScriptFrame{}, false
}

type ScriptRunner struct {
	runningScripts []*ScriptInstance
}

func NewScriptRunner() *ScriptRunner {
	return &ScriptRunner{
		runningScripts: make([]*ScriptInstance, 0),
	}
}

func (s *ScriptRunner) RunScript(dataDir string, scriptName string, condFuncs map[string]govaluate.ExpressionFunction) {
	script := LoadScript(dataDir, scriptName, condFuncs)
	runningScript := &ScriptInstance{
		script:       script,
		currentFrame: 0,
	}
	s.runningScripts = append(s.runningScripts, runningScript)
}

func (s *ScriptRunner) OnTurn() {
	s.runningScripts = fxtools.FilterSlice(s.runningScripts, func(instance *ScriptInstance) bool {
		return !instance.IsDone()
	})

	for _, instance := range s.runningScripts {
		if endFrame, isValid := instance.HasReachedOutcome(); isValid {
			instance.Run(endFrame)
		} else if instance.CanRunCurrentFrame() {
			instance.RunCurrentFrame()
		}
	}
}
