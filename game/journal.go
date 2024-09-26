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

type QuestState uint8

const (
	QuestUnknown QuestState = iota
	QuestStarted
	QuestInProgress
	QuestCompleted
)

type Quest struct {
	// fixed
	Identifier  string
	DisplayName string
	RewardInXP  int

	Starters []*JournalEntry
	Progress []*JournalEntry
	Outcomes []*NamedJournalEntry

	// changing
	CurrentState  QuestState
	StartIndex    int
	ProgressIndex int
	Outcome       string
}

func (q *Quest) GetActiveEntry() *JournalEntry {
	switch q.CurrentState {
	case QuestStarted:
		return q.getStart()
	case QuestInProgress:
		return q.getProgress()
	case QuestCompleted:
		return q.getOutcome()
	}
	return nil
}

func (q *Quest) getOutcome() *JournalEntry {
	for _, entry := range q.Outcomes {
		if entry.Identifier == q.Outcome {
			return entry.JournalEntry
		}
	}
	return nil
}

func (q *Quest) getStart() *JournalEntry {
	return q.Starters[q.StartIndex]
}

func (q *Quest) getProgress() *JournalEntry {
	return q.Progress[q.ProgressIndex]
}

func (q *Quest) HasNewState() (newState bool, setFlagName string) {
	switch q.CurrentState {
	case QuestCompleted:
		return false, ""
	case QuestUnknown:
		startIndex := firstValidEntry(q.Starters)
		if startIndex != -1 {
			q.StartIndex = startIndex
			q.CurrentState = QuestStarted
			return true, fmt.Sprintf("QuestStarted(%s)", q.Identifier)
		}
	case QuestStarted:
		progressIndex := firstValidEntry(q.Progress)
		if progressIndex != -1 && progressIndex != q.ProgressIndex {
			q.ProgressIndex = progressIndex
			q.CurrentState = QuestInProgress
			return true, fmt.Sprintf("QuestInProgress(%s)", q.Identifier)
		}
		fallthrough
	case QuestInProgress:
		endIndex := firstValidEntry(q.Outcomes)
		if endIndex != -1 && q.Outcome != q.Outcomes[endIndex].Identifier {
			q.Outcome = q.Outcomes[endIndex].Identifier
			q.CurrentState = QuestCompleted
			return true, fmt.Sprintf("QuestCompleted(%s, %s)", q.Identifier, q.Outcome)
		}
	}
	return false, ""
}

func firstValidEntry[T Conditional](starters []T) int {
	for index, entry := range starters {
		if entry.GetCondition() == nil {
			return index
		}
		evaluate, err := entry.GetCondition().Evaluate(nil)
		if err != nil {
			panic(err)
		}
		if evaluate.(bool) {
			return index
		}
	}
	return -1
}

type Conditional interface {
	GetCondition() *govaluate.EvaluableExpression
}
type JournalEntry struct {
	Entry         string
	HasBeenViewed bool
	Condition     *govaluate.EvaluableExpression
}

func (j *JournalEntry) IsValid() bool {
	return j.Condition != nil && j.Entry != ""
}
func (j *JournalEntry) GetCondition() *govaluate.EvaluableExpression {
	return j.Condition
}

type NamedJournalEntry struct {
	*JournalEntry
	Identifier string
}

func (n *NamedJournalEntry) IsValid() bool {
	return n.JournalEntry.IsValid() && n.Identifier != ""
}

func (n *NamedJournalEntry) ToRecord(prefix string) recfile.Record {
	return append(n.JournalEntry.ToRecord(prefix), recfile.Field{Name: prefix + "_id", Value: n.Identifier})
}
func (j *JournalEntry) ToRecord(prefix string) recfile.Record {
	return recfile.Record{
		recfile.Field{Name: prefix + "_text", Value: j.Entry},
		recfile.Field{Name: prefix + "_cond", Value: j.Condition.String()},
		recfile.Field{Name: prefix + "_viewed", Value: recfile.BoolStr(j.HasBeenViewed)},
	}
}

type Journal struct {
	quests           map[string][]*Quest
	onJournalChanged func()
	incrementFlag    func(string)
}

// NewJournal Is used during initialization of the game state.
func NewJournal(io io.ReadCloser, fMap map[string]govaluate.ExpressionFunction) *Journal {
	j := &Journal{quests: make(map[string][]*Quest)}
	j.AddEntriesFromSource("default", io, fMap)
	io.Close()
	return j
}

func (j *Journal) AddEntriesFromSource(context string, reader io.Reader, fMap map[string]govaluate.ExpressionFunction) {
	records := recfile.Read(reader)
	for _, record := range records {
		entry := NewQuestFromRecord(record, fMap)
		j.quests[context] = append(j.quests[context], entry)
	}
}

func NewQuestFromRecord(record recfile.Record, fMap map[string]govaluate.ExpressionFunction) *Quest {
	quest := &Quest{
		StartIndex:    -1,
		ProgressIndex: -1,
	}

	lastStateParsed := QuestUnknown

	currentJournalEntry := &NamedJournalEntry{
		JournalEntry: &JournalEntry{},
	}

	stateFromFieldName := func(fieldName string) QuestState {
		part := strings.Split(strings.ToLower(fieldName), "_")[0]
		switch part {
		case "start":
			return QuestStarted
		case "prog":
			return QuestInProgress
		case "end":
			return QuestCompleted
		}
		return QuestUnknown
	}

	commitCurrentEntry := func(entry *NamedJournalEntry) {
		switch lastStateParsed {
		case QuestStarted:
			quest.Starters = append(quest.Starters, entry.JournalEntry)
		case QuestInProgress:
			quest.Progress = append(quest.Progress, entry.JournalEntry)
		case QuestCompleted:
			quest.Outcomes = append(quest.Outcomes, entry)
		}
	}

	for _, field := range record {
		switch strings.ToLower(field.Name) {
		case "id":
			quest.Identifier = field.Value
		case "name":
			quest.DisplayName = field.Value
		case "xp":
			quest.RewardInXP = recfile.StrInt(field.Value)
		default:
			if !strings.Contains(field.Name, "_") {
				panic(fmt.Sprintf("Unknown field name: %s", field.Name))
			}

			if (lastStateParsed == QuestCompleted && currentJournalEntry.IsValid()) ||
				(lastStateParsed != QuestCompleted && currentJournalEntry.JournalEntry.IsValid()) {
				commitCurrentEntry(currentJournalEntry)
				currentJournalEntry = &NamedJournalEntry{
					JournalEntry: &JournalEntry{},
				}
			}
			lastStateParsed = stateFromFieldName(field.Name)

			if strings.HasSuffix(field.Name, "_text") {
				currentJournalEntry.Entry = strings.TrimSpace(field.Value)
			} else if strings.HasSuffix(field.Name, "_cond") {
				currentJournalEntry.Condition, _ = govaluate.NewEvaluableExpressionWithFunctions(field.Value, fMap)
			} else if strings.HasSuffix(field.Name, "_id") {
				currentJournalEntry.Identifier = field.Value
			}
		}
	}
	if currentJournalEntry.IsValid() || currentJournalEntry.JournalEntry.IsValid() {
		commitCurrentEntry(currentJournalEntry)
	}
	return quest
}

func (q *Quest) ToRecord() recfile.Record {
	result := make([]recfile.Field, 0, 3+len(q.Starters)*2+len(q.Progress)*2+len(q.Outcomes)*3)
	result = append(result, recfile.Field{Name: "ID", Value: q.Identifier})
	result = append(result, recfile.Field{Name: "Name", Value: q.DisplayName})
	result = append(result, recfile.Field{Name: "XP", Value: recfile.IntStr(q.RewardInXP)})

	for _, entry := range q.Starters {
		result = append(result, entry.ToRecord("start_")...)
	}
	for _, entry := range q.Progress {
		result = append(result, entry.ToRecord("prog_")...)
	}
	for _, entry := range q.Outcomes {
		result = append(result, entry.ToRecord("end_")...)
	}
	return result

}

// NewJournalFromRecords creates a new Journal from a map of records. This is used during game state loading.
func NewJournalFromRecords(records map[string][]recfile.Record, fMap map[string]govaluate.ExpressionFunction) *Journal {
	j := &Journal{quests: make(map[string][]*Quest)}
	for context, recordList := range records {
		for _, record := range recordList {
			entry := NewQuestFromRecord(record, fMap)
			j.quests[context] = append(j.quests[context], entry)
		}
	}
	return j
}

func (j *Journal) SetChangeHandler(onJournalChanged func()) {
	j.onJournalChanged = onJournalChanged
}

func (j *Journal) Update() []Reward {
	sawChanges := false
	var rewards []Reward
	for context, _ := range j.quests {
		// update our active items
		for _, quest := range j.quests[context] {
			hasNewState, flagToIncrement := quest.HasNewState()
			if hasNewState {
				sawChanges = true
				if quest.CurrentState == QuestCompleted {
					rewards = append(rewards, Reward{XP: quest.RewardInXP, Text: quest.DisplayName})
					j.incrementFlag(fmt.Sprintf("QuestCompleted(%s)", quest.Identifier))
				}
				j.incrementFlag(flagToIncrement)
			}
		}
	}

	if sawChanges && j.onJournalChanged != nil {
		j.onJournalChanged()
	}

	return rewards
}
func (j *Journal) GetEntriesForViewing(context string) []string {
	var result []string
	for _, quest := range j.getActiveQuests(context) {
		entry := quest.GetActiveEntry()
		if entry == nil {
			continue
		}
		var entryString string
		if entry.HasBeenViewed {
			entryString = entry.Entry
		} else {
			fgCode := textiles.RGBAToFgColorCode(color.RGBA{R: 20, G: 240, B: 20, A: 255})
			coloredString := fmt.Sprintf("%s%s[-:-:-]", fgCode, entry.Entry)
			entry.HasBeenViewed = true
			entryString = coloredString
		}
		result = append(result, fmt.Sprintf("## [green]%s[-] ##\n%s", quest.DisplayName, entryString))
	}
	return result
}

func (j *Journal) getActiveQuests(context string) []*Quest {
	var result []*Quest
	for _, q := range j.quests[context] {
		if q.CurrentState != QuestUnknown {
			result = append(result, q)
		}
	}
	return result
}

func (j *Journal) ToRecords() map[string][]recfile.Record {
	records := make(map[string][]recfile.Record)
	for context, quests := range j.quests {
		records[context] = make([]recfile.Record, len(quests))
		for i, q := range quests {
			records[context][i] = q.ToRecord()
		}
	}
	return records
}

func (j *Journal) SetIncrementFlagHandler(increment func(key string)) {
	j.incrementFlag = increment
}
