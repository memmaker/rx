package game

import (
	"cmp"
	"fmt"
	"github.com/memmaker/go/fxtools"
	"github.com/memmaker/go/recfile"
	"path"
	"slices"
	"strings"
	"time"
)

type DayOfWeek int

const (
	Sunday DayOfWeek = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday

	Everyday
)

func (d DayOfWeek) DayBefore() DayOfWeek {
	if d == Sunday {
		return Saturday
	}
	return d - 1
}

func WeekdayFromString(s string) DayOfWeek {
	switch strings.ToLower(s) {
	case "Monday":
		return Monday
	case "tuesday":
		return Tuesday
	case "wednesday":
		return Wednesday
	case "thursday":
		return Thursday
	case "friday":
		return Friday
	case "saturday":
		return Saturday
	case "sunday":
		return Sunday
	case "every_day":
		return Everyday
	}
	return Monday
}

type TimeSlot struct {
	Day                DayOfWeek
	Time               time.Time
	MapName            string
	Location           string
	ActionFormatString string
	ObservationFlag    string
	UseTransition      string
}

func (s TimeSlot) Description() string {
	return s.ActionFormatString
}

type SlotID struct {
	Day   DayOfWeek
	Index int
}
type Schedule struct {
	Slots      map[DayOfWeek][]TimeSlot
	LastSlotID SlotID
	SourceFile string
}

func clockAsSeconds(t time.Time) int {
	hour, m, sec := t.Clock()
	return hour*3600 + m*60 + sec
}

func (s *Schedule) MoveToNextTimeSlot(now time.Time) (TimeSlot, bool) {
	// return the time slot that is in the past and the nearest time slot
	// but is not the LastSlotID

	day := DayOfWeek(now.Weekday())
	allSlots := s.slotsForDay(day)

	if len(allSlots) == 0 {
		return TimeSlot{}, false
	}

	lastSlotID := -1
	if s.LastSlotID.Day == day {
		lastSlotID = s.LastSlotID.Index
	}

	lastSlot := allSlots[len(allSlots)-1] // first in time

	if clockAsSeconds(lastSlot.Time) > clockAsSeconds(now) {
		var slotsBefore []TimeSlot
		var dayBefore DayOfWeek
		for len(slotsBefore) == 0 {
			dayBefore = day.DayBefore()
			slotsBefore = s.slotsForDay(dayBefore)
		}
		if s.LastSlotID.Day == dayBefore {
			lastSlotID = s.LastSlotID.Index
		}

		if 0 < lastSlotID || lastSlotID == -1 {
			// last slot of the day before
			lastSlot = slotsBefore[0]
			s.LastSlotID = SlotID{Day: dayBefore, Index: 0}
			return lastSlot, true
		}
		return TimeSlot{}, false
	}

	for index, slot := range allSlots {
		if clockAsSeconds(slot.Time) <= clockAsSeconds(now) && (index < lastSlotID || lastSlotID == -1) {
			s.LastSlotID = SlotID{Day: day, Index: index}
			return slot, true
		}
	}

	return TimeSlot{}, false
}

func (s *Schedule) slotsForDay(day DayOfWeek) []TimeSlot {
	allSlots := append(s.Slots[day], s.Slots[Everyday]...)

	if len(allSlots) == 0 {
		return nil
	}

	slices.SortStableFunc(allSlots, func(i, j TimeSlot) int {
		return cmp.Compare(clockAsSeconds(j.Time), clockAsSeconds(i.Time))
	})
	return allSlots
}

func (s *Schedule) CurrentTimeSlot() TimeSlot {
	if s.LastSlotID.Index == -1 {
		return TimeSlot{}
	}
	allSlots := s.slotsForDay(s.LastSlotID.Day)
	if len(allSlots) == 0 {
		return TimeSlot{}
	}
	return allSlots[s.LastSlotID.Index]
}

func (s *Schedule) String() string {
	return fmt.Sprintf("%s (%d)", s.SourceFile, len(s.Slots))
}

func NewScheduleFromFile(filename string) *Schedule {
	schedule := &Schedule{
		SourceFile: path.Base(filename),
		Slots:      make(map[DayOfWeek][]TimeSlot),
		LastSlotID: SlotID{Index: -1},
	}
	file := fxtools.MustOpen(filename)
	defer file.Close()
	records, _ := recfile.ReadMulti(file)
	slots := records["slots"]

	for _, slot := range slots {
		newSlot := NewTimeSlotFromRecord(slot)
		schedule.Slots[newSlot.Day] = append(schedule.Slots[newSlot.Day], newSlot)
	}

	return schedule
}

func NewTimeSlotFromRecord(slot recfile.Record) TimeSlot {
	timeSlot := TimeSlot{}
	for _, field := range slot {
		switch strings.ToLower(field.Name) {
		case "day":
			timeSlot.Day = WeekdayFromString(field.Value)
		case "time":
			timeSlot.Time, _ = time.Parse("15:04", field.Value)
		case "location":
			timeSlot.Location = field.Value
		case "map":
			timeSlot.MapName = field.Value
		case "transition":
			timeSlot.UseTransition = field.Value
		case "description":
			timeSlot.ActionFormatString = field.Value
		case "observation_flag":
			timeSlot.ObservationFlag = field.Value
		}
	}
	return timeSlot

}
