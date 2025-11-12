package storage

import (
	"context"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
)

type TransactionStorage interface {
	GetStatistics(ctx context.Context, req GetStatisticsRequest) ([]models.PeriodStats, error)
	GetTransactionsForForecast(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error)
	GetCategoryStatsByPeriods(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.CategoryPeriodStats, error)
	GetRecurringPatterns(ctx context.Context, userID string) ([]models.RecurringPattern, error)
}

type GetStatisticsRequest struct {
	UserID    string
	StartDate time.Time
	EndDate   time.Time
	GroupBy   models.TimePeriod
}
