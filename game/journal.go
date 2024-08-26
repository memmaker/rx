package game

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/memmaker/go/recfile"
	"github.com/memmaker/go/textiles"
	"image/color"
	"io"
	"strings"
)

type JournalEntry struct {
	Entry         string
	Condition     *govaluate.EvaluableExpression
	IsActive      bool
	HasBeenViewed bool
}

func NewJournalEntry(record recfile.Record, fMap map[string]govaluate.ExpressionFunction) *JournalEntry {
	var entry JournalEntry
	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "text":
			entry.Entry = strings.TrimSpace(field.Value)
		case "cond":
			expression, err := govaluate.NewEvaluableExpressionWithFunctions(field.Value, fMap)
			if err != nil {
				panic(err)
			}
			entry.Condition = expression
		}
	}
	return &entry
}

type Journal struct {
	entries          map[string][]*JournalEntry
	onJournalChanged func()
}

func NewJournal(io io.ReadCloser, fMap map[string]govaluate.ExpressionFunction) *Journal {
	j := &Journal{entries: make(map[string][]*JournalEntry)}
	j.AddEntriesFromSource("", io, fMap)
	io.Close()
	return j
}

func (j *Journal) SetChangeHandler(onJournalChanged func()) {
	j.onJournalChanged = onJournalChanged
}

func (j *Journal) AddEntriesFromSource(context string, reader io.Reader, fMap map[string]govaluate.ExpressionFunction) {
	records := recfile.Read(reader)
	for _, record := range records {
		entry := NewJournalEntry(record, fMap)
		j.entries[context] = append(j.entries[context], entry)
	}
}

func (j *Journal) OnFlagsChanged() {
	sawChanges := false
	for context, _ := range j.entries {
		// update our active items
		for i, entry := range j.entries[context] {
			if entry.Condition != nil {
				evaluate, err := entry.Condition.Evaluate(nil)
				if err != nil {
					panic(err)
				}
				newResult := evaluate.(bool)
				oldResult := j.entries[context][i].IsActive
				if newResult != oldResult {
					j.entries[context][i].IsActive = newResult
					sawChanges = true
				}
			}
		}
	}

	if sawChanges && j.onJournalChanged != nil {
		j.onJournalChanged()
	}
}
func (j *Journal) GetEntries(context string, params map[string]interface{}) []*JournalEntry {
	var result []*JournalEntry
	for _, entry := range j.entries[context] {
		if entry.Condition == nil {
			result = append(result, entry)
		} else {
			evaluate, err := entry.Condition.Evaluate(params)
			if err != nil {
				panic(err)
			}
			if err == nil && evaluate.(bool) {
				result = append(result, entry)
			}
		}
	}
	return result
}

func (j *Journal) GetEntriesForViewing(context string) []string {
	var result []string
	for _, entry := range j.getActiveEntries(context) {
		if entry.HasBeenViewed {
			result = append(result, entry.Entry)
		} else {
			fgCode := textiles.RGBAToFgColorCode(color.RGBA{R: 20, G: 240, B: 20, A: 255})
			coloredString := fmt.Sprintf("%s%s[-:-:-]", fgCode, entry.Entry)
			result = append(result, coloredString)
			entry.HasBeenViewed = true
		}
	}
	return result
}

func (j *Journal) getActiveEntries(context string) []*JournalEntry {
	var result []*JournalEntry
	for _, entry := range j.entries[context] {
		if entry.IsActive {
			result = append(result, entry)
		}
	}
	return result
}
