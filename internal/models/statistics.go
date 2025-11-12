package models

import "time"

type PeriodStats struct {
	PeriodStart time.Time
	PeriodEnd   time.Time
	Income      int64
	Expense     int64
	Balance     int64
	Categories  []CategoryStats
}

type CategoryStats struct {
	CategoryID  string
	TotalAmount int64
}

type TimePeriod string

const (
	TimePeriodMonth   TimePeriod = "MONTH"
	TimePeriodQuarter TimePeriod = "QUARTER"
	TimePeriodYear    TimePeriod = "YEAR"
)
