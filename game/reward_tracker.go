package game

import (
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/recfile"
	"strings"
)

type Reward struct {
	XP   int
	Text string
}

func NewConditionalReward(record recfile.Record, fMap map[string]govaluate.ExpressionFunction) Reward {
	var entry Reward
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "xp":
			entry.XP = recfile.StrInt(field.Value)
		case "text":
			entry.Text = field.Value
		}
	}
	return entry
}
