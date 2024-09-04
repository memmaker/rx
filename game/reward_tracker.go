package game

import (
	"cmp"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/recfile"
	"io"
	"slices"
	"strings"
)

type ConditionalReward struct {
	XP        int
	Condition *govaluate.EvaluableExpression
	Text      string
}

func NewConditionalReward(record recfile.Record, fMap map[string]govaluate.ExpressionFunction) ConditionalReward {
	var entry ConditionalReward
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "cond":
			entry.Condition, _ = govaluate.NewEvaluableExpressionWithFunctions(field.Value, fMap)
		case "xp":
			entry.XP = recfile.StrInt(field.Value)
		case "text":
			entry.Text = field.Value
		}
	}
	return entry
}

type RewardTracker struct {
	entries        []ConditionalReward
	alreadyAwarded map[int]bool // indices of entries that have already been awarded
}

func NewRewardTracker(io io.ReadCloser, fMap map[string]govaluate.ExpressionFunction) *RewardTracker {
	j := &RewardTracker{entries: make([]ConditionalReward, 0), alreadyAwarded: make(map[int]bool)}
	j.AddEntriesFromSource(io, fMap)
	io.Close()
	return j
}

func (j *RewardTracker) GetNewRewards(params map[string]interface{}) []ConditionalReward {
	var newRewards []ConditionalReward
	for i, entry := range j.entries {
		if j.alreadyAwarded[i] {
			continue
		}
		if entry.Condition != nil {
			result, _ := entry.Condition.Evaluate(params)
			if result.(bool) {
				newRewards = append(newRewards, entry)
				j.alreadyAwarded[i] = true
			}
		}
	}
	return newRewards
}

func (j *RewardTracker) AddEntriesFromSource(reader io.Reader, fMap map[string]govaluate.ExpressionFunction) {
	records := recfile.Read(reader)
	for _, record := range records {
		entry := NewConditionalReward(record, fMap)
		j.entries = append(j.entries, entry)
	}
}

func (j *RewardTracker) SetRewardsReceived(received []int) {
	for _, idx := range received {
		j.alreadyAwarded[idx] = true
	}
}

func (j *RewardTracker) GetRewardsReceived() []int {
	var received []int
	for idx, _ := range j.alreadyAwarded {
		received = append(received, idx)
	}
	slices.SortStableFunc(received, func(i, j int) int { return cmp.Compare(i, j) })
	return received
}
