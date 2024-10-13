package game

import (
    "cmp"
    "fmt"
    "slices"
    "strings"
    "time"
)

type PointInTime struct {
    Turns int
    Time  time.Time
}

func (t PointInTime) AddDuration(duration time.Duration) PointInTime {
    return PointInTime{
        Turns: t.Turns,
        Time:  t.Time.Add(duration),
    }
}

func (t PointInTime) AddDurationAndTurn(duration time.Duration) PointInTime {
    return PointInTime{
        Turns: t.Turns + 1,
        Time:  t.Time.Add(duration),
    }
}

func (t PointInTime) WithTurns(turns int) PointInTime {
    return PointInTime{
        Turns: turns,
        Time:  t.Time,
    }
}

func (t PointInTime) WithTime(time time.Time) PointInTime {
    return PointInTime{
        Turns: t.Turns,
        Time:  time,
    }
}

func (t PointInTime) TurnsSince(inTime PointInTime) int {
    return t.Turns - inTime.Turns
}

func (t PointInTime) MinutesSince(inTime PointInTime) int {
    return int(t.Time.Sub(inTime.Time).Minutes())
}

func (t PointInTime) HoursSince(inTime PointInTime) int {
    return int(t.Time.Sub(inTime.Time).Hours())
}

func (t PointInTime) DaysSince(inTime PointInTime) int {
    return int(t.Time.Sub(inTime.Time).Hours() / 24)
}

type TimeTracker map[string]PointInTime

func (t TimeTracker) String() string {
    if len(t) == 0 {
        return "No time tracked"
    }
    out := make([]string, len(t))
    i := 0
    for key, value := range t {
        out[i] = fmt.Sprintf("%s: %s", key, value.Time.Format("2006-01-02 15:04"))
        i++
    }
    slices.SortStableFunc(out, func(i, j string) int {
        return cmp.Compare(i, j)
    })
    return strings.Join(out, "\n")
}
