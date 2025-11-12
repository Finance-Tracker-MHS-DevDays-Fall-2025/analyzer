package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/storage"
)

func TestGetStatistics_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

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

	service := NewAnalyzerService(mockStorage, logger)

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
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger)

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
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger)

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
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger)

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

	mockStorage := storage.NewMockStorage()
	mockStorage.GetStatisticsFunc = func(ctx context.Context, req storage.GetStatisticsRequest) ([]models.PeriodStats, error) {
		return []models.PeriodStats{}, nil
	}

	service := NewAnalyzerService(mockStorage, logger)

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

	service := NewAnalyzerService(mockStorage, logger)

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
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger)

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

	service := NewAnalyzerService(mockStorage, logger)

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
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger)

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

	mockStorage := storage.NewMockStorage()
	mockStorage.GetTransactionsForForecastFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
		return []models.PeriodStats{
			{Income: 100000, Expense: 50000},
			{Income: 95000, Expense: 48000},
		}, nil
	}

	service := NewAnalyzerService(mockStorage, logger)

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
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger)

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
	mockStorage := storage.NewMockStorage()
	service := NewAnalyzerService(mockStorage, logger)

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

	service := NewAnalyzerService(mockStorage, logger)

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

	service := NewAnalyzerService(mockStorage, logger)

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
