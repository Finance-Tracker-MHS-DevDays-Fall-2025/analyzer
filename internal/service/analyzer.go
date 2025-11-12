package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/storage"
)

type AnalyzerService struct {
	storage storage.TransactionStorage
	logger  *slog.Logger
}

func NewAnalyzerService(storage storage.TransactionStorage, logger *slog.Logger) *AnalyzerService {
	return &AnalyzerService{
		storage: storage,
		logger:  logger.With("component", "analyzer_service"),
	}
}

func (s *AnalyzerService) GetStatistics(ctx context.Context, userID string, startDate, endDate time.Time, groupBy models.TimePeriod) ([]models.PeriodStats, int64, int64, error) {
	if userID == "" {
		return nil, 0, 0, fmt.Errorf("user_id is required")
	}

	if startDate.IsZero() || endDate.IsZero() {
		return nil, 0, 0, fmt.Errorf("start_date and end_date are required")
	}

	if startDate.After(endDate) {
		return nil, 0, 0, fmt.Errorf("start_date must be before end_date")
	}

	if groupBy == "" {
		groupBy = models.TimePeriodMonth
	}

	req := storage.GetStatisticsRequest{
		UserID:    userID,
		StartDate: startDate,
		EndDate:   endDate,
		GroupBy:   groupBy,
	}

	periods, err := s.storage.GetStatistics(ctx, req)
	if err != nil {
		s.logger.Error("failed to get statistics", "error", err, "user_id", userID)
		return nil, 0, 0, fmt.Errorf("failed to get statistics: %w", err)
	}

	totalIncome := int64(0)
	totalExpense := int64(0)

	for _, period := range periods {
		totalIncome += period.Income
		totalExpense += period.Expense
	}

	s.logger.Info("statistics calculated",
		"user_id", userID,
		"periods", len(periods),
		"total_income", totalIncome,
		"total_expense", totalExpense,
	)

	return periods, totalIncome, totalExpense, nil
}

func (s *AnalyzerService) GetForecast(ctx context.Context, userID string, period models.TimePeriod, periodsAhead int) ([]models.PeriodStats, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	if periodsAhead <= 0 {
		periodsAhead = 1
	}

	if periodsAhead > 12 {
		return nil, fmt.Errorf("periods_ahead cannot exceed 12")
	}

	if period == "" {
		period = models.TimePeriodMonth
	}

	lookbackPeriods := 6
	now := time.Now()
	currentPeriodStart := truncateToPeriodStart(now, period)
	startDate := calculateStartDate(currentPeriodStart, period, lookbackPeriods)

	historicalData, err := s.storage.GetTransactionsForForecast(ctx, userID, startDate, lookbackPeriods, period)
	if err != nil {
		s.logger.Error("failed to get historical data", "error", err, "user_id", userID)
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	if len(historicalData) < 2 {
		return nil, fmt.Errorf("insufficient historical data for forecast (need at least 2 periods)")
	}

	forecasts := s.calculateWMAForecast(historicalData, periodsAhead, period)

	s.logger.Info("forecast calculated",
		"user_id", userID,
		"periods_ahead", periodsAhead,
		"historical_periods", len(historicalData),
	)

	return forecasts, nil
}

func (s *AnalyzerService) calculateWMAForecast(historical []models.PeriodStats, periodsAhead int, period models.TimePeriod) []models.PeriodStats {
	n := len(historical)
	if n > 6 {
		n = 6
	}

	weights := make([]float64, n)
	totalWeight := 0.0
	for i := 0; i < n; i++ {
		weights[i] = float64(n - i)
		totalWeight += weights[i]
	}

	avgIncome := 0.0
	avgExpense := 0.0

	for i := 0; i < n; i++ {
		weight := weights[i] / totalWeight
		avgIncome += float64(historical[i].Income) * weight
		avgExpense += float64(historical[i].Expense) * weight
	}

	forecasts := make([]models.PeriodStats, periodsAhead)
	lastPeriod := historical[0].PeriodStart

	for i := 0; i < periodsAhead; i++ {
		periodStart := calculateNextPeriod(lastPeriod, period, i+1)
		periodEnd := calculatePeriodEnd(periodStart, period)

		forecasts[i] = models.PeriodStats{
			PeriodStart: periodStart,
			PeriodEnd:   periodEnd,
			Income:      int64(avgIncome),
			Expense:     int64(avgExpense),
			Balance:     int64(avgIncome) - int64(avgExpense),
			Categories:  []models.CategoryStats{},
		}
	}

	return forecasts
}

func truncateToPeriodStart(t time.Time, period models.TimePeriod) time.Time {
	year, month, _ := t.Date()
	switch period {
	case models.TimePeriodMonth:
		return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	case models.TimePeriodQuarter:
		quarterMonth := ((int(month)-1)/3)*3 + 1
		return time.Date(year, time.Month(quarterMonth), 1, 0, 0, 0, 0, t.Location())
	case models.TimePeriodYear:
		return time.Date(year, 1, 1, 0, 0, 0, 0, t.Location())
	default:
		return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
	}
}

func calculateStartDate(periodStart time.Time, period models.TimePeriod, lookbackPeriods int) time.Time {
	switch period {
	case models.TimePeriodMonth:
		return periodStart.AddDate(0, -lookbackPeriods, 0)
	case models.TimePeriodQuarter:
		return periodStart.AddDate(0, -lookbackPeriods*3, 0)
	case models.TimePeriodYear:
		return periodStart.AddDate(-lookbackPeriods, 0, 0)
	default:
		return periodStart.AddDate(0, -lookbackPeriods, 0)
	}
}

func calculateNextPeriod(base time.Time, period models.TimePeriod, offset int) time.Time {
	switch period {
	case models.TimePeriodMonth:
		return base.AddDate(0, offset, 0)
	case models.TimePeriodQuarter:
		return base.AddDate(0, offset*3, 0)
	case models.TimePeriodYear:
		return base.AddDate(offset, 0, 0)
	default:
		return base.AddDate(0, offset, 0)
	}
}

func calculatePeriodEnd(start time.Time, period models.TimePeriod) time.Time {
	switch period {
	case models.TimePeriodMonth:
		return start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	case models.TimePeriodQuarter:
		return start.AddDate(0, 3, 0).Add(-time.Nanosecond)
	case models.TimePeriodYear:
		return start.AddDate(1, 0, 0).Add(-time.Nanosecond)
	default:
		return start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	}
}
