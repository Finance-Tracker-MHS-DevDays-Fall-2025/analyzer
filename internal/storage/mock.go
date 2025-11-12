package storage

import (
	"context"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
)

type MockStorage struct {
	GetStatisticsFunc              func(ctx context.Context, req GetStatisticsRequest) ([]models.PeriodStats, error)
	GetTransactionsForForecastFunc func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error)
}

func NewMockStorage() *MockStorage {
	return &MockStorage{}
}

func (m *MockStorage) GetStatistics(ctx context.Context, req GetStatisticsRequest) ([]models.PeriodStats, error) {
	if m.GetStatisticsFunc != nil {
		return m.GetStatisticsFunc(ctx, req)
	}
	return []models.PeriodStats{}, nil
}

func (m *MockStorage) GetTransactionsForForecast(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
	if m.GetTransactionsForForecastFunc != nil {
		return m.GetTransactionsForForecastFunc(ctx, userID, startDate, periods, groupBy)
	}
	return []models.PeriodStats{}, nil
}
