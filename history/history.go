package history

import "time"

type Timeframe string

const (
	TimeframeDay   Timeframe = "day"
	TimeframeWeek  Timeframe = "week"
	TimeframeMonth Timeframe = "month"
	TimeframeYear  Timeframe = "year"
)

func (t Timeframe) GetNextStartTime(current time.Time) time.Time {
	switch t {
	case TimeframeDay:
		return current.AddDate(0, 0, 1)
	case TimeframeWeek:
		return current.AddDate(0, 0, 7)
	case TimeframeMonth:
		return current.AddDate(0, 1, 0)
	case TimeframeYear:
		return current.AddDate(1, 0, 0)
	default:
		return time.Now()
	}
}

func (t Timeframe) DefaultFirstStartTime() *time.Time {
	now := time.Now()
	switch t {
	case TimeframeDay:
		monthAgo := now.AddDate(0, -1, 0)
		start := time.Date(monthAgo.Year(), monthAgo.Month(), monthAgo.Day(), 0, 0, 0, 0, time.Local)
		return &start
	case TimeframeWeek:
		twelveWeeksAgo := now.AddDate(0, 0, -12*7)
		start := time.Date(twelveWeeksAgo.Year(), twelveWeeksAgo.Month(), twelveWeeksAgo.Day(), 0, 0, 0, 0, time.Local)
		return &start
	default:
		return nil
	}
}
