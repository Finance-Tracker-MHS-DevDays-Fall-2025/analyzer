package models

import "time"

type RecurringPattern struct {
	MCC             string
	MedianAmount    int64
	AvgIntervalDays float64
	LastOccurrence  time.Time
}

type RecurringPayment struct {
	MCC           string
	TypicalAmount int64
	ExpectedDate  time.Time
}
