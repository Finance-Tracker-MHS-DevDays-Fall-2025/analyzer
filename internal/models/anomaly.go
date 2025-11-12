package models

import "time"

type CategoryPeriodStats struct {
	PeriodStart time.Time
	CategoryID  string
	Amount      int64
}

type CategoryAnomaly struct {
	MCC             string
	ActualAmount    int64
	ExpectedAmount  int64
	DeviationAmount int64
}
