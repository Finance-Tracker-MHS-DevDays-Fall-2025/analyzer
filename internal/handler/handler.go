package handler

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/service"
	pb "github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/analyzer"
	pbcommon "github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/pkg/api/proto/common"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AnalyzerHandler struct {
	pb.UnimplementedAnalyzerServiceServer
	service *service.AnalyzerService
	logger  *slog.Logger
}

func NewAnalyzerHandler(
	service *service.AnalyzerService,
	logger *slog.Logger,
) *AnalyzerHandler {
	return &AnalyzerHandler{
		service: service,
		logger:  logger.With("component", "analyzer_handler"),
	}
}

func (h *AnalyzerHandler) GetStatistics(ctx context.Context, req *pb.GetStatisticsRequest) (*pb.GetStatisticsResponse, error) {
	h.logger.Info("GetStatistics called", "user_id", req.UserId)

	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.StartDate == nil || req.EndDate == nil {
		return nil, fmt.Errorf("start_date and end_date are required")
	}

	if !req.StartDate.IsValid() || !req.EndDate.IsValid() {
		return nil, fmt.Errorf("invalid timestamp format")
	}

	groupBy := parseTimePeriod(req.GroupBy)

	periods, totalIncome, totalExpense, err := h.service.GetStatistics(
		ctx,
		req.UserId,
		req.StartDate.AsTime(),
		req.EndDate.AsTime(),
		groupBy,
	)
	if err != nil {
		h.logger.Error("failed to get statistics", "error", err, "user_id", req.UserId)
		return nil, err
	}

	return &pb.GetStatisticsResponse{
		TotalIncome:  &pbcommon.Money{Amount: totalIncome, Currency: "RUB"},
		TotalExpense: &pbcommon.Money{Amount: totalExpense, Currency: "RUB"},
		PeriodData:   convertPeriodsToPB(periods),
	}, nil
}

func (h *AnalyzerHandler) GetForecast(ctx context.Context, req *pb.GetForecastRequest) (*pb.GetForecastResponse, error) {
	h.logger.Info("GetForecast called", "user_id", req.UserId)

	period := parseTimePeriod(req.Period)

	forecasts, err := h.service.GetForecast(ctx, req.UserId, period, int(req.PeriodsAhead))
	if err != nil {
		h.logger.Error("failed to get forecast", "error", err, "user_id", req.UserId)
		return nil, err
	}

	return &pb.GetForecastResponse{
		Forecasts: convertForecastsToPB(forecasts),
	}, nil
}

func parseTimePeriod(pbPeriod pbcommon.TimePeriod) models.TimePeriod {
	switch pbPeriod {
	case pbcommon.TimePeriod_TIME_PERIOD_MONTH:
		return models.TimePeriodMonth
	case pbcommon.TimePeriod_TIME_PERIOD_QUARTER:
		return models.TimePeriodQuarter
	case pbcommon.TimePeriod_TIME_PERIOD_YEAR:
		return models.TimePeriodYear
	default:
		return models.TimePeriodMonth
	}
}

func convertPeriodsToPB(periods []models.PeriodStats) []*pb.PeriodBalance {
	result := make([]*pb.PeriodBalance, 0, len(periods))

	for _, p := range periods {
		result = append(result, &pb.PeriodBalance{
			PeriodStart:       timestamppb.New(p.PeriodStart),
			PeriodEnd:         timestamppb.New(p.PeriodEnd),
			Income:            &pbcommon.Money{Amount: p.Income, Currency: "RUB"},
			Expense:           &pbcommon.Money{Amount: p.Expense, Currency: "RUB"},
			Balance:           &pbcommon.Money{Amount: p.Balance, Currency: "RUB"},
			CategoryBreakdown: convertCategoriesToPB(p.Categories),
		})
	}

	return result
}

func convertCategoriesToPB(categories []models.CategoryStats) []*pb.CategorySpending {
	result := make([]*pb.CategorySpending, 0, len(categories))

	for _, c := range categories {
		result = append(result, &pb.CategorySpending{
			CategoryId:  c.CategoryID,
			TotalAmount: &pbcommon.Money{Amount: c.TotalAmount, Currency: "RUB"},
		})
	}

	return result
}

func convertForecastsToPB(forecasts []models.PeriodStats) []*pb.Forecast {
	result := make([]*pb.Forecast, 0, len(forecasts))

	for _, f := range forecasts {
		result = append(result, &pb.Forecast{
			PeriodStart:       timestamppb.New(f.PeriodStart),
			PeriodEnd:         timestamppb.New(f.PeriodEnd),
			ExpectedIncome:    &pbcommon.Money{Amount: f.Income, Currency: "RUB"},
			ExpectedExpense:   &pbcommon.Money{Amount: f.Expense, Currency: "RUB"},
			ExpectedBalance:   &pbcommon.Money{Amount: f.Balance, Currency: "RUB"},
			CategoryBreakdown: convertCategoriesToPB(f.Categories),
		})
	}

	return result
}
