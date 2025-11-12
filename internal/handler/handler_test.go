package handler

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/service"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/storage"
	pb "github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/analyzer"
	pbcommon "github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestGetStatistics_Handler_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mockStorage := storage.NewMockStorage()
	mockStorage.GetStatisticsFunc = func(ctx context.Context, req storage.GetStatisticsRequest) ([]models.PeriodStats, error) {
		return []models.PeriodStats{
			{
				PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
				Income:      150000,
				Expense:     80000,
				Balance:     70000,
				Categories: []models.CategoryStats{
					{CategoryID: "5411", TotalAmount: 50000},
					{CategoryID: "5812", TotalAmount: 30000},
				},
			},
		}, nil
	}

	analyzerService := service.NewAnalyzerService(mockStorage, logger)
	handler := NewAnalyzerHandler(analyzerService, logger)

	req := &pb.GetStatisticsRequest{
		UserId:    "user-123",
		StartDate: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:   timestamppb.New(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
		GroupBy:   pbcommon.TimePeriod_TIME_PERIOD_MONTH,
	}

	resp, err := handler.GetStatistics(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.TotalIncome.Amount != 150000 {
		t.Errorf("expected total income 150000, got %d", resp.TotalIncome.Amount)
	}

	if resp.TotalExpense.Amount != 80000 {
		t.Errorf("expected total expense 80000, got %d", resp.TotalExpense.Amount)
	}

	if len(resp.PeriodData) != 1 {
		t.Fatalf("expected 1 period, got %d", len(resp.PeriodData))
	}

	period := resp.PeriodData[0]

	if period.Income.Amount != 150000 {
		t.Errorf("expected period income 150000, got %d", period.Income.Amount)
	}

	if period.Expense.Amount != 80000 {
		t.Errorf("expected period expense 80000, got %d", period.Expense.Amount)
	}

	if period.Balance.Amount != 70000 {
		t.Errorf("expected period balance 70000, got %d", period.Balance.Amount)
	}

	if len(period.CategoryBreakdown) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(period.CategoryBreakdown))
	}

	if period.CategoryBreakdown[0].CategoryId != "5411" {
		t.Errorf("expected category 5411, got %s", period.CategoryBreakdown[0].CategoryId)
	}

	if period.CategoryBreakdown[0].TotalAmount.Amount != 50000 {
		t.Errorf("expected category amount 50000, got %d", period.CategoryBreakdown[0].TotalAmount.Amount)
	}
}

func TestGetStatistics_Handler_MultiplePeriods(t *testing.T) {
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
				Categories:  []models.CategoryStats{},
			},
			{
				PeriodStart: time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 2, 29, 23, 59, 59, 0, time.UTC),
				Income:      110000,
				Expense:     55000,
				Balance:     55000,
				Categories:  []models.CategoryStats{},
			},
			{
				PeriodStart: time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
				PeriodEnd:   time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
				Income:      120000,
				Expense:     60000,
				Balance:     60000,
				Categories:  []models.CategoryStats{},
			},
		}, nil
	}

	analyzerService := service.NewAnalyzerService(mockStorage, logger)
	handler := NewAnalyzerHandler(analyzerService, logger)

	req := &pb.GetStatisticsRequest{
		UserId:    "user-123",
		StartDate: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:   timestamppb.New(time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC)),
		GroupBy:   pbcommon.TimePeriod_TIME_PERIOD_MONTH,
	}

	resp, err := handler.GetStatistics(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.TotalIncome.Amount != 330000 {
		t.Errorf("expected total income 330000, got %d", resp.TotalIncome.Amount)
	}

	if resp.TotalExpense.Amount != 165000 {
		t.Errorf("expected total expense 165000, got %d", resp.TotalExpense.Amount)
	}

	if len(resp.PeriodData) != 3 {
		t.Fatalf("expected 3 periods, got %d", len(resp.PeriodData))
	}
}

func TestGetStatistics_Handler_CurrencyIsRUB(t *testing.T) {
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
				Categories:  []models.CategoryStats{{CategoryID: "5411", TotalAmount: 30000}},
			},
		}, nil
	}

	analyzerService := service.NewAnalyzerService(mockStorage, logger)
	handler := NewAnalyzerHandler(analyzerService, logger)

	req := &pb.GetStatisticsRequest{
		UserId:    "user-123",
		StartDate: timestamppb.New(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)),
		EndDate:   timestamppb.New(time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC)),
		GroupBy:   pbcommon.TimePeriod_TIME_PERIOD_MONTH,
	}

	resp, err := handler.GetStatistics(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.TotalIncome.Currency != "RUB" {
		t.Errorf("expected currency RUB, got %s", resp.TotalIncome.Currency)
	}

	if resp.TotalExpense.Currency != "RUB" {
		t.Errorf("expected currency RUB, got %s", resp.TotalExpense.Currency)
	}

	if resp.PeriodData[0].Income.Currency != "RUB" {
		t.Errorf("expected period currency RUB, got %s", resp.PeriodData[0].Income.Currency)
	}

	if resp.PeriodData[0].CategoryBreakdown[0].TotalAmount.Currency != "RUB" {
		t.Errorf("expected category currency RUB, got %s", resp.PeriodData[0].CategoryBreakdown[0].TotalAmount.Currency)
	}
}

func TestGetForecast_Handler_Success(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	mockStorage := storage.NewMockStorage()
	mockStorage.GetTransactionsForForecastFunc = func(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
		return []models.PeriodStats{
			{Income: 100000, Expense: 50000, Balance: 50000},
			{Income: 95000, Expense: 48000, Balance: 47000},
			{Income: 90000, Expense: 45000, Balance: 45000},
		}, nil
	}

	analyzerService := service.NewAnalyzerService(mockStorage, logger)
	handler := NewAnalyzerHandler(analyzerService, logger)

	req := &pb.GetForecastRequest{
		UserId:       "user-123",
		Period:       pbcommon.TimePeriod_TIME_PERIOD_MONTH,
		PeriodsAhead: 3,
	}

	resp, err := handler.GetForecast(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Forecasts) != 3 {
		t.Fatalf("expected 3 forecasts, got %d", len(resp.Forecasts))
	}

	forecast := resp.Forecasts[0]

	if forecast.ExpectedIncome.Amount == 0 {
		t.Error("expected non-zero expected income")
	}

	if forecast.ExpectedExpense.Amount == 0 {
		t.Error("expected non-zero expected expense")
	}

	if forecast.ExpectedBalance.Amount != forecast.ExpectedIncome.Amount-forecast.ExpectedExpense.Amount {
		t.Error("expected balance = income - expense")
	}

	if forecast.ExpectedIncome.Currency != "RUB" {
		t.Errorf("expected currency RUB, got %s", forecast.ExpectedIncome.Currency)
	}
}

func TestGetForecast_Handler_QuarterlyPeriod(t *testing.T) {
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

	analyzerService := service.NewAnalyzerService(mockStorage, logger)
	handler := NewAnalyzerHandler(analyzerService, logger)

	req := &pb.GetForecastRequest{
		UserId:       "user-123",
		Period:       pbcommon.TimePeriod_TIME_PERIOD_QUARTER,
		PeriodsAhead: 2,
	}

	resp, err := handler.GetForecast(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Forecasts) != 2 {
		t.Fatalf("expected 2 forecasts, got %d", len(resp.Forecasts))
	}
}

func TestGetForecast_Handler_YearlyPeriod(t *testing.T) {
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

	analyzerService := service.NewAnalyzerService(mockStorage, logger)
	handler := NewAnalyzerHandler(analyzerService, logger)

	req := &pb.GetForecastRequest{
		UserId:       "user-123",
		Period:       pbcommon.TimePeriod_TIME_PERIOD_YEAR,
		PeriodsAhead: 2,
	}

	resp, err := handler.GetForecast(context.Background(), req)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(resp.Forecasts) != 2 {
		t.Fatalf("expected 2 forecasts, got %d", len(resp.Forecasts))
	}
}

func TestParseTimePeriod_AllValues(t *testing.T) {
	tests := []struct {
		name     string
		input    pbcommon.TimePeriod
		expected models.TimePeriod
	}{
		{"Month", pbcommon.TimePeriod_TIME_PERIOD_MONTH, models.TimePeriodMonth},
		{"Quarter", pbcommon.TimePeriod_TIME_PERIOD_QUARTER, models.TimePeriodQuarter},
		{"Year", pbcommon.TimePeriod_TIME_PERIOD_YEAR, models.TimePeriodYear},
		{"Unspecified", pbcommon.TimePeriod_TIME_PERIOD_UNSPECIFIED, models.TimePeriodMonth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseTimePeriod(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestConvertPeriodsToPB_EmptyCategories(t *testing.T) {
	periods := []models.PeriodStats{
		{
			PeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:   time.Date(2024, 1, 31, 23, 59, 59, 0, time.UTC),
			Income:      100000,
			Expense:     50000,
			Balance:     50000,
			Categories:  []models.CategoryStats{},
		},
	}

	result := convertPeriodsToPB(periods)

	if len(result) != 1 {
		t.Fatalf("expected 1 period, got %d", len(result))
	}

	if len(result[0].CategoryBreakdown) != 0 {
		t.Errorf("expected 0 categories, got %d", len(result[0].CategoryBreakdown))
	}
}

func TestConvertCategoriesToPB_MultipleCategories(t *testing.T) {
	categories := []models.CategoryStats{
		{CategoryID: "5411", TotalAmount: 50000},
		{CategoryID: "5812", TotalAmount: 30000},
		{CategoryID: "uncategorized", TotalAmount: 20000},
	}

	result := convertCategoriesToPB(categories)

	if len(result) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(result))
	}

	if result[0].CategoryId != "5411" {
		t.Errorf("expected category 5411, got %s", result[0].CategoryId)
	}

	if result[0].TotalAmount.Amount != 50000 {
		t.Errorf("expected amount 50000, got %d", result[0].TotalAmount.Amount)
	}

	if result[2].CategoryId != "uncategorized" {
		t.Errorf("expected category uncategorized, got %s", result[2].CategoryId)
	}
}
