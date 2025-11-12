package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/config"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/storage"
)

func getDefaultTestConfig() *config.AnalyticsConfig {
	return &config.AnalyticsConfig{
		Forecast: config.ForecastConfig{
			LookbackPeriods: 6,
			MaxPeriodsAhead: 12,
		},
		Anomaly: config.AnomalyConfig{
			LookbackPeriods:      6,
			DeviationThreshold:   50.0,
			NewCategoryThreshold: 50000,
		},
		Recurring: config.RecurringConfig{
			LookbackMonths:    6,
			MinOccurrences:    3,
			IntervalMinDays:   25,
			IntervalMaxDays:   35,
			DateDeviationDays: 3,
			PredictionDays:    30,
		},
	}
}

func TestGetStatistics_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetStatisticsFunc = func(ctx context.Context, req storage.GetStatisticsRequest) ([]models.PeriodStats, error) {
		return []models.PeriodStats{
			{
				PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				Income:      100000,
				Expense:     50000,
				Balance:     50000,
				Categories: []models.CategoryStats{
					{CategoryID: "5411", TotalAmount: 30000},
					{CategoryID: "5812", TotalAmount: 20000},
				},
			},
			{
				PeriodStart: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC),
				Income:      120000,
				Expense:     60000,
				Balance:     60000,
				Categories: []models.CategoryStats{
					{CategoryID: "5411", TotalAmount: 35000},
					{CategoryID: "5812", TotalAmount: 25000},
				},
			},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC)

	periods, totalIncome, totalExpense, err := service.GetStatistics(
		context.Background(),
		"user-123",
		startDate,
		endDate,
		models.TimePeriodMonth,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(periods) != 2 {
		t.Fatalf("expected 2 periods, got %d", len(periods))
	}

	if totalIncome != 220000 {
		t.Errorf("expected total income 220000, got %d", totalIncome)
	}

	if totalExpense != 110000 {
		t.Errorf("expected total expense 110000, got %d", totalExpense)
	}

	if periods[0].Balance != 50000 {
		t.Errorf("expected first period balance 50000, got %d", periods[0].Balance)
	}

	if len(periods[0].Categories) != 2 {
		t.Errorf("expected 2 categories in first period, got %d", len(periods[0].Categories))
	}
}

func TestGetStatistics_EmptyUserID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, _, _, err := service.GetStatistics(
		context.Background(),
		"",
		time.Now(),
		time.Now(),
		models.TimePeriodMonth,
	)

	if err == nil {
		t.Fatal("expected error for empty user_id, got nil")
	}

	expectedMsg := "user_id is required"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetStatistics_InvalidDateRange(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	startDate := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	_, _, _, err := service.GetStatistics(
		context.Background(),
		"user-123",
		startDate,
		endDate,
		models.TimePeriodMonth,
	)

	if err == nil {
		t.Fatal("expected error for invalid date range, got nil")
	}

	expectedMsg := "start_date must be before end_date"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetStatistics_ZeroDates(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, _, _, err := service.GetStatistics(
		context.Background(),
		"user-123",
		time.Time{},
		time.Time{},
		models.TimePeriodMonth,
	)

	if err == nil {
		t.Fatal("expected error for zero dates, got nil")
	}

	expectedMsg := "start_date and end_date are required"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetStatistics_NoData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetStatisticsFunc = func(ctx context.Context, req storage.GetStatisticsRequest) ([]models.PeriodStats, error) {
		return []models.PeriodStats{}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)

	periods, totalIncome, totalExpense, err := service.GetStatistics(
		context.Background(),
		"user-123",
		startDate,
		endDate,
		models.TimePeriodMonth,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(periods) != 0 {
		t.Errorf("expected 0 periods, got %d", len(periods))
	}

	if totalIncome != 0 {
		t.Errorf("expected total income 0, got %d", totalIncome)
	}

	if totalExpense != 0 {
		t.Errorf("expected total expense 0, got %d", totalExpense)
	}
}

func TestGetForecast_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetTransactionsForForecastFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
		return []models.PeriodStats{
			{
				PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 6, 30, 23, 59, 59, 0, time.UTC),
				Income:      100000,
				Expense:     50000,
				Balance:     50000,
			},
			{
				PeriodStart: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 5, 31, 23, 59, 59, 0, time.UTC),
				Income:      95000,
				Expense:     48000,
				Balance:     47000,
			},
			{
				PeriodStart: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 4, 30, 23, 59, 59, 0, time.UTC),
				Income:      90000,
				Expense:     45000,
				Balance:     45000,
			},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	forecasts, err := service.GetForecast(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
		3,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(forecasts) != 3 {
		t.Fatalf("expected 3 forecast periods, got %d", len(forecasts))
	}

	if forecasts[0].Income == 0 {
		t.Error("expected non-zero forecast income")
	}

	if forecasts[0].Expense == 0 {
		t.Error("expected non-zero forecast expense")
	}

	if forecasts[0].Balance != forecasts[0].Income-forecasts[0].Expense {
		t.Error("balance should equal income - expense")
	}

	for i := 1; i < len(forecasts); i++ {
		if forecasts[i].Income != forecasts[0].Income {
			t.Errorf("all forecasts should have same income, period %d differs", i)
		}
		if forecasts[i].Expense != forecasts[0].Expense {
			t.Errorf("all forecasts should have same expense, period %d differs", i)
		}
	}
}

func TestGetForecast_EmptyUserID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, err := service.GetForecast(
		context.Background(),
		"",
		models.TimePeriodMonth,
		3,
	)

	if err == nil {
		t.Fatal("expected error for empty user_id, got nil")
	}
}

func TestGetForecast_InsufficientData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetTransactionsForForecastFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
		return []models.PeriodStats{
			{
				PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				Income:      100000,
				Expense:     50000,
				Balance:     50000,
			},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, err := service.GetForecast(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
		3,
	)

	if err == nil {
		t.Fatal("expected error for insufficient data, got nil")
	}

	expectedMsg := "insufficient historical data for forecast (need at least 2 periods)"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetForecast_TooManyPeriodsAhead(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, err := service.GetForecast(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
		13,
	)

	if err == nil {
		t.Fatal("expected error for too many periods ahead, got nil")
	}

	expectedMsg := "periods_ahead cannot exceed 12"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetForecast_ZeroPeriodsAhead(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetTransactionsForForecastFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
		return []models.PeriodStats{
			{Income: 100000, Expense: 50000},
			{Income: 95000, Expense: 48000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	forecasts, err := service.GetForecast(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
		0,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(forecasts) != 1 {
		t.Errorf("expected 1 forecast (default), got %d", len(forecasts))
	}
}

func TestCalculateWMAForecast_WeightedCorrectly(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	historical := []models.PeriodStats{
		{
			PeriodStart: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
			Income:      120000,
			Expense:     60000,
		},
		{
			PeriodStart: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
			Income:      100000,
			Expense:     50000,
		},
		{
			PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			Income:      80000,
			Expense:     40000,
		},
	}

	forecasts := service.calculateWMAForecast(historical, 2, models.TimePeriodMonth)

	if len(forecasts) != 2 {
		t.Fatalf("expected 2 forecasts, got %d", len(forecasts))
	}

	weightedIncome := (120000.0*3.0 + 100000.0*2.0 + 80000.0*1.0) / 6.0
	weightedExpense := (60000.0*3.0 + 50000.0*2.0 + 40000.0*1.0) / 6.0
	expectedIncome := int64(weightedIncome)
	expectedExpense := int64(weightedExpense)

	if forecasts[0].Income < expectedIncome-1000 || forecasts[0].Income > expectedIncome+1000 {
		t.Errorf("expected income around %d, got %d", expectedIncome, forecasts[0].Income)
	}

	if forecasts[0].Expense < expectedExpense-1000 || forecasts[0].Expense > expectedExpense+1000 {
		t.Errorf("expected expense around %d, got %d", expectedExpense, forecasts[0].Expense)
	}

	if forecasts[0].Income != forecasts[1].Income {
		t.Error("all WMA forecasts should have the same income")
	}
}

func TestCalculateWMAForecast_MaxSixPeriods(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	historical := make([]models.PeriodStats, 10)
	for i := 0; i < 10; i++ {
		historical[i] = models.PeriodStats{
			PeriodStart: time.Date(2024, time.Month(10-i), 1, 0, 0, 0, 0, time.UTC),
			Income:      100000,
			Expense:     50000,
		}
	}

	forecasts := service.calculateWMAForecast(historical, 1, models.TimePeriodMonth)

	if len(forecasts) != 1 {
		t.Fatalf("expected 1 forecast, got %d", len(forecasts))
	}

	if forecasts[0].Income != 100000 {
		t.Errorf("expected income 100000 (all periods same), got %d", forecasts[0].Income)
	}
}

func TestGetForecast_QuarterlyPeriod(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetTransactionsForForecastFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
		if groupBy != models.TimePeriodQuarter {
			t.Error("expected TimePeriodQuarter")
		}
		return []models.PeriodStats{
			{Income: 300000, Expense: 150000},
			{Income: 280000, Expense: 140000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	forecasts, err := service.GetForecast(
		context.Background(),
		"user-123",
		models.TimePeriodQuarter,
		2,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(forecasts) != 2 {
		t.Fatalf("expected 2 forecasts, got %d", len(forecasts))
	}
}

func TestGetForecast_YearlyPeriod(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetTransactionsForForecastFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
		if groupBy != models.TimePeriodYear {
			t.Error("expected TimePeriodYear")
		}
		return []models.PeriodStats{
			{Income: 1200000, Expense: 600000},
			{Income: 1150000, Expense: 580000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	forecasts, err := service.GetForecast(
		context.Background(),
		"user-123",
		models.TimePeriodYear,
		2,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(forecasts) != 2 {
		t.Fatalf("expected 2 forecasts, got %d", len(forecasts))
	}
}

func TestGetAnomalies_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetCategoryStatsByPeriodsFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.CategoryPeriodStats, error) {
		return []models.CategoryPeriodStats{
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 150000},
			{PeriodStart: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 80000},
			{PeriodStart: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 75000},
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5812", Amount: 40000},
			{PeriodStart: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5812", Amount: 38000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	anomalies, err := service.GetAnomalies(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(anomalies) == 0 {
		t.Error("expected at least one anomaly")
	}

	found := false
	for _, a := range anomalies {
		if a.MCC == "5411" && a.DeviationAmount > 0 {
			found = true
			if a.ActualAmount != 150000 {
				t.Errorf("expected actual amount 150000, got %d", a.ActualAmount)
			}
			if a.ExpectedAmount == 0 {
				t.Error("expected non-zero expected amount")
			}
		}
	}

	if !found {
		t.Error("expected anomaly for MCC 5411")
	}
}

func TestGetAnomalies_EmptyUserID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, err := service.GetAnomalies(
		context.Background(),
		"",
		models.TimePeriodMonth,
	)

	if err == nil {
		t.Fatal("expected error for empty user_id, got nil")
	}
}

func TestGetAnomalies_InsufficientData(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetCategoryStatsByPeriodsFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.CategoryPeriodStats, error) {
		return []models.CategoryPeriodStats{
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 100000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, err := service.GetAnomalies(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
	)

	if err == nil {
		t.Fatal("expected error for insufficient data, got nil")
	}

	expectedMsg := "insufficient data for anomaly detection (need at least 2 periods)"
	if err.Error() != expectedMsg {
		t.Errorf("expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGetAnomalies_NewCategory(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetCategoryStatsByPeriodsFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.CategoryPeriodStats, error) {
		return []models.CategoryPeriodStats{
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 100000},
			{PeriodStart: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 95000},
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "9999", Amount: 60000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	anomalies, err := service.GetAnomalies(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	found := false
	for _, a := range anomalies {
		if a.MCC == "9999" {
			found = true
			if a.ExpectedAmount != 0 {
				t.Errorf("expected zero expected amount for new category, got %d", a.ExpectedAmount)
			}
			if a.ActualAmount != 60000 {
				t.Errorf("expected actual amount 60000, got %d", a.ActualAmount)
			}
			if a.DeviationAmount != 60000 {
				t.Errorf("expected deviation amount 60000, got %d", a.DeviationAmount)
			}
		}
	}

	if !found {
		t.Error("expected anomaly for new category MCC 9999")
	}
}

func TestGetAnomalies_BelowThreshold(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetCategoryStatsByPeriodsFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.CategoryPeriodStats, error) {
		return []models.CategoryPeriodStats{
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 105000},
			{PeriodStart: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 100000},
			{PeriodStart: time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 95000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	anomalies, err := service.GetAnomalies(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(anomalies) != 0 {
		t.Errorf("expected no anomalies (deviation below threshold), got %d", len(anomalies))
	}
}

func TestGetAnomalies_SortedByDeviation(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetCategoryStatsByPeriodsFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.CategoryPeriodStats, error) {
		return []models.CategoryPeriodStats{
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 200000},
			{PeriodStart: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5411", Amount: 100000},
			{PeriodStart: time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5812", Amount: 120000},
			{PeriodStart: time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC), CategoryID: "5812", Amount: 50000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	anomalies, err := service.GetAnomalies(
		context.Background(),
		"user-123",
		models.TimePeriodMonth,
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(anomalies) < 2 {
		t.Fatalf("expected at least 2 anomalies, got %d", len(anomalies))
	}

	for i := 1; i < len(anomalies); i++ {
		if anomalies[i].DeviationAmount > anomalies[i-1].DeviationAmount {
			t.Error("anomalies should be sorted by deviation amount in descending order")
		}
	}
}

func TestGetUpcomingRecurring_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	now := time.Now()
	lastOccurrence := now.AddDate(0, 0, -25)

	mockStorage := storage.NewMockStorage()
	mockStorage.GetRecurringPatternsFunc = func(ctx context.Context, userID string) ([]models.RecurringPattern, error) {
		return []models.RecurringPattern{
			{
				MCC:             "5411",
				MedianAmount:    80000,
				AvgIntervalDays: 30,
				LastOccurrence:  lastOccurrence,
			},
			{
				MCC:             "5812",
				MedianAmount:    50000,
				AvgIntervalDays: 30,
				LastOccurrence:  now.AddDate(0, 0, -10),
			},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	payments, err := service.GetUpcomingRecurring(
		context.Background(),
		"user-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(payments) == 0 {
		t.Error("expected at least one upcoming payment")
	}

	found := false
	for _, p := range payments {
		if p.MCC == "5411" {
			found = true
			if p.TypicalAmount != 80000 {
				t.Errorf("expected typical amount 80000, got %d", p.TypicalAmount)
			}
			expectedDate := lastOccurrence.AddDate(0, 0, 30)
			if !p.ExpectedDate.Equal(expectedDate) {
				t.Errorf("expected date %v, got %v", expectedDate, p.ExpectedDate)
			}
		}
	}

	if !found {
		t.Error("expected upcoming payment for MCC 5411")
	}
}

func TestGetUpcomingRecurring_EmptyUserID(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	_, err := service.GetUpcomingRecurring(
		context.Background(),
		"",
	)

	if err == nil {
		t.Fatal("expected error for empty user_id, got nil")
	}
}

func TestGetUpcomingRecurring_NoPatterns(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetRecurringPatternsFunc = func(ctx context.Context, userID string) ([]models.RecurringPattern, error) {
		return []models.RecurringPattern{}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	payments, err := service.GetUpcomingRecurring(
		context.Background(),
		"user-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(payments) != 0 {
		t.Errorf("expected 0 payments, got %d", len(payments))
	}
}

func TestGetUpcomingRecurring_OnlyPastPayments(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	now := time.Now()
	lastOccurrence := now.AddDate(0, 0, -60)

	mockStorage := storage.NewMockStorage()
	mockStorage.GetRecurringPatternsFunc = func(ctx context.Context, userID string) ([]models.RecurringPattern, error) {
		return []models.RecurringPattern{
			{
				MCC:             "5411",
				MedianAmount:    80000,
				AvgIntervalDays: 30,
				LastOccurrence:  lastOccurrence,
			},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	payments, err := service.GetUpcomingRecurring(
		context.Background(),
		"user-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(payments) != 0 {
		t.Errorf("expected 0 payments (all in the past), got %d", len(payments))
	}
}

func TestGetUpcomingRecurring_OnlyFarFuturePayments(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	now := time.Now()
	lastOccurrence := now.AddDate(0, 0, -5)

	mockStorage := storage.NewMockStorage()
	mockStorage.GetRecurringPatternsFunc = func(ctx context.Context, userID string) ([]models.RecurringPattern, error) {
		return []models.RecurringPattern{
			{
				MCC:             "5411",
				MedianAmount:    80000,
				AvgIntervalDays: 60,
				LastOccurrence:  lastOccurrence,
			},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	payments, err := service.GetUpcomingRecurring(
		context.Background(),
		"user-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(payments) != 0 {
		t.Errorf("expected 0 payments (beyond prediction window), got %d", len(payments))
	}
}

func TestGetUpcomingRecurring_SortedByDate(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()

	now := time.Now()

	mockStorage := storage.NewMockStorage()
	mockStorage.GetRecurringPatternsFunc = func(ctx context.Context, userID string) ([]models.RecurringPattern, error) {
		return []models.RecurringPattern{
			{
				MCC:             "5411",
				MedianAmount:    80000,
				AvgIntervalDays: 20,
				LastOccurrence:  now.AddDate(0, 0, -10),
			},
			{
				MCC:             "5812",
				MedianAmount:    50000,
				AvgIntervalDays: 15,
				LastOccurrence:  now.AddDate(0, 0, -10),
			},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger, cfg)

	payments, err := service.GetUpcomingRecurring(
		context.Background(),
		"user-123",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(payments) < 2 {
		t.Fatalf("expected at least 2 payments, got %d", len(payments))
	}

	for i := 1; i < len(payments); i++ {
		if payments[i].ExpectedDate.Before(payments[i-1].ExpectedDate) {
			t.Error("payments should be sorted by expected date in ascending order")
		}
	}
}

func TestCalculateWMAByCategory_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	periodData := map[time.Time]map[string]int64{
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC): {"5411": 120000, "5812": 60000},
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC): {"5411": 100000, "5812": 50000},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC): {"5411": 80000, "5812": 40000},
	}

	periods := []time.Time{
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	expected := service.calculateWMAByCategory(periodData, periods)

	if len(expected) != 2 {
		t.Errorf("expected 2 categories, got %d", len(expected))
	}

	weightedMCC5411 := (120000.0*3.0 + 100000.0*2.0 + 80000.0*1.0) / 6.0
	if expected["5411"] < int64(weightedMCC5411)-1000 || expected["5411"] > int64(weightedMCC5411)+1000 {
		t.Errorf("expected MCC 5411 around %d, got %d", int64(weightedMCC5411), expected["5411"])
	}

	weightedMCC5812 := (60000.0*3.0 + 50000.0*2.0 + 40000.0*1.0) / 6.0
	if expected["5812"] < int64(weightedMCC5812)-1000 || expected["5812"] > int64(weightedMCC5812)+1000 {
		t.Errorf("expected MCC 5812 around %d, got %d", int64(weightedMCC5812), expected["5812"])
	}
}

func TestCalculateWMAByCategory_EmptyPeriods(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	periodData := map[time.Time]map[string]int64{}
	periods := []time.Time{}

	expected := service.calculateWMAByCategory(periodData, periods)

	if len(expected) != 0 {
		t.Errorf("expected 0 categories, got %d", len(expected))
	}
}

func TestCalculateWMAByCategory_MissingCategoryInPeriod(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cfg := getDefaultTestConfig()
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger, cfg)

	periodData := map[time.Time]map[string]int64{
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC): {"5411": 120000},
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC): {"5411": 100000, "5812": 50000},
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC): {"5411": 80000},
	}

	periods := []time.Time{
		time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	expected := service.calculateWMAByCategory(periodData, periods)

	if len(expected) != 2 {
		t.Errorf("expected 2 categories, got %d", len(expected))
	}

	if _, exists := expected["5812"]; !exists {
		t.Error("expected MCC 5812 to be present")
	}

	weightedMCC5812 := (0.0*3.0 + 50000.0*2.0 + 0.0*1.0) / 6.0
	if expected["5812"] < int64(weightedMCC5812)-1000 || expected["5812"] > int64(weightedMCC5812)+1000 {
		t.Errorf("expected MCC 5812 around %d, got %d", int64(weightedMCC5812), expected["5812"])
	}
}

func TestTruncateToPeriodStart_Month(t *testing.T) {
	input := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)
	expected := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	result := truncateToPeriodStart(input, models.TimePeriodMonth)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestTruncateToPeriodStart_Quarter(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected time.Time
	}{
		{time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2024, 4, 15, 0, 0, 0, 0, time.UTC), time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2024, 7, 15, 0, 0, 0, 0, time.UTC), time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)},
		{time.Date(2024, 10, 15, 0, 0, 0, 0, time.UTC), time.Date(2024, 10, 1, 0, 0, 0, 0, time.UTC)},
	}

	for _, tt := range tests {
		result := truncateToPeriodStart(tt.input, models.TimePeriodQuarter)
		if !result.Equal(tt.expected) {
			t.Errorf("for input %v expected %v, got %v", tt.input, tt.expected, result)
		}
	}
}

func TestTruncateToPeriodStart_Year(t *testing.T) {
	input := time.Date(2024, 6, 15, 14, 30, 45, 0, time.UTC)
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	result := truncateToPeriodStart(input, models.TimePeriodYear)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculateStartDate_Month(t *testing.T) {
	periodStart := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	result := calculateStartDate(periodStart, models.TimePeriodMonth, 3)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculateStartDate_Quarter(t *testing.T) {
	periodStart := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	result := calculateStartDate(periodStart, models.TimePeriodQuarter, 2)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculateStartDate_Year(t *testing.T) {
	periodStart := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	result := calculateStartDate(periodStart, models.TimePeriodYear, 2)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculateNextPeriod_Month(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC)
	result := calculateNextPeriod(base, models.TimePeriodMonth, 3)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculateNextPeriod_Quarter(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC)
	result := calculateNextPeriod(base, models.TimePeriodQuarter, 2)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculateNextPeriod_Year(t *testing.T) {
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	expected := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	result := calculateNextPeriod(base, models.TimePeriodYear, 2)

	if !result.Equal(expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculatePeriodEnd_Month(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	result := calculatePeriodEnd(start, models.TimePeriodMonth)

	expected := time.Date(2024, 1, 31, 23, 59, 59, 999999999, time.UTC)
	if result.Year() != expected.Year() || result.Month() != expected.Month() || result.Day() != expected.Day() {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculatePeriodEnd_Quarter(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	result := calculatePeriodEnd(start, models.TimePeriodQuarter)

	expected := time.Date(2024, 3, 31, 23, 59, 59, 999999999, time.UTC)
	if result.Year() != expected.Year() || result.Month() != expected.Month() || result.Day() != expected.Day() {
		t.Errorf("expected %v, got %v", expected, result)
	}
}

func TestCalculatePeriodEnd_Year(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	result := calculatePeriodEnd(start, models.TimePeriodYear)

	expected := time.Date(2024, 12, 31, 23, 59, 59, 999999999, time.UTC)
	if result.Year() != expected.Year() || result.Month() != expected.Month() || result.Day() != expected.Day() {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
