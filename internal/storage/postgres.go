package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/config"
	"github.com/Finance-Tracker-MHS-DevDays-Fall-2025/analyzer/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
	cfg  *config.RecurringConfig
}

func NewPostgresStorage(pool *pgxpool.Pool, cfg *config.RecurringConfig) *PostgresStorage {
	return &PostgresStorage{
		pool: pool,
		cfg:  cfg,
	}
}

func (s *PostgresStorage) GetStatistics(ctx context.Context, req GetStatisticsRequest) ([]models.PeriodStats, error) {
	truncFunc := getTruncFunction(req.GroupBy)

	query := `
		WITH user_transactions AS (
			SELECT 
				t.type,
				t.amount,
				t.currency,
				t.mcc,
				t.created_at
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			WHERE a.user_id = $1
				AND t.created_at >= $2
				AND t.created_at <= $3
				AND t.type IN ('INCOME', 'EXPENSE')
		),
		period_aggregates AS (
			SELECT 
				DATE_TRUNC($4, created_at) as period_start,
				type,
				SUM(amount) as total_amount
			FROM user_transactions
			GROUP BY DATE_TRUNC($4, created_at), type
		),
		category_aggregates AS (
			SELECT 
				DATE_TRUNC($4, created_at) as period_start,
				COALESCE(mcc::TEXT, 'uncategorized') as category_id,
				SUM(amount) as total_amount
			FROM user_transactions
			WHERE type = 'EXPENSE'
			GROUP BY DATE_TRUNC($4, created_at), mcc
		)
		SELECT 
			pa.period_start,
			COALESCE(SUM(CASE WHEN pa.type = 'INCOME' THEN pa.total_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN pa.type = 'EXPENSE' THEN pa.total_amount ELSE 0 END), 0) as expense,
			COALESCE(ca.category_id, '') as category_id,
			COALESCE(ca.total_amount, 0) as category_amount
		FROM period_aggregates pa
		LEFT JOIN category_aggregates ca ON pa.period_start = ca.period_start
		GROUP BY pa.period_start, ca.category_id, ca.total_amount
		ORDER BY pa.period_start, ca.total_amount DESC NULLS LAST
	`

	rows, err := s.pool.Query(ctx, query, req.UserID, req.StartDate, req.EndDate, truncFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to query statistics: %w", err)
	}
	defer rows.Close()

	periodsMap := make(map[time.Time]*models.PeriodStats)
	var periodKeys []time.Time

	for rows.Next() {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		var periodStart time.Time
		var income, expense int64
		var categoryID string
		var categoryAmount int64

		if err := rows.Scan(&periodStart, &income, &expense, &categoryID, &categoryAmount); err != nil {
			return nil, fmt.Errorf("failed to scan period row: %w", err)
		}

		period, exists := periodsMap[periodStart]
		if !exists {
			period = &models.PeriodStats{
				PeriodStart: periodStart,
				PeriodEnd:   calculatePeriodEnd(periodStart, req.GroupBy),
				Income:      income,
				Expense:     expense,
				Balance:     income - expense,
				Categories:  []models.CategoryStats{},
			}
			periodsMap[periodStart] = period
			periodKeys = append(periodKeys, periodStart)
		}

		if categoryID != "" && categoryAmount > 0 {
			period.Categories = append(period.Categories, models.CategoryStats{
				CategoryID:  categoryID,
				TotalAmount: categoryAmount,
			})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating period rows: %w", err)
	}

	periods := make([]models.PeriodStats, 0, len(periodKeys))
	for _, key := range periodKeys {
		periods = append(periods, *periodsMap[key])
	}

	return periods, nil
}

func (s *PostgresStorage) getCategoryBreakdown(ctx context.Context, userID string, startDate, endDate time.Time) ([]models.CategoryStats, error) {
	query := `
		SELECT 
			COALESCE(t.mcc::TEXT, 'uncategorized') as category_id,
			SUM(t.amount) as total_amount
		FROM transactions t
		JOIN accounts a ON t.account_id = a.id
		WHERE a.user_id = $1
			AND t.created_at >= $2
			AND t.created_at < $3
			AND t.type = 'EXPENSE'
		GROUP BY t.mcc
		ORDER BY total_amount DESC
	`

	rows, err := s.pool.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.CategoryStats

	for rows.Next() {
		var cat models.CategoryStats
		if err := rows.Scan(&cat.CategoryID, &cat.TotalAmount); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}
		categories = append(categories, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category rows: %w", err)
	}

	return categories, nil
}

func (s *PostgresStorage) GetTransactionsForForecast(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.PeriodStats, error) {
	truncFunc := getTruncFunction(groupBy)

	query := `
		WITH user_transactions AS (
			SELECT 
				t.type,
				t.amount,
				t.created_at
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			WHERE a.user_id = $1
				AND t.created_at >= $2
				AND t.type IN ('INCOME', 'EXPENSE')
		),
		period_aggregates AS (
			SELECT 
				DATE_TRUNC($3, created_at) as period_start,
				type,
				SUM(amount) as total_amount
			FROM user_transactions
			GROUP BY DATE_TRUNC($3, created_at), type
		)
		SELECT 
			pa.period_start,
			COALESCE(SUM(CASE WHEN pa.type = 'INCOME' THEN pa.total_amount ELSE 0 END), 0) as income,
			COALESCE(SUM(CASE WHEN pa.type = 'EXPENSE' THEN pa.total_amount ELSE 0 END), 0) as expense
		FROM period_aggregates pa
		GROUP BY pa.period_start
		ORDER BY pa.period_start DESC
		LIMIT $4
	`

	rows, err := s.pool.Query(ctx, query, userID, startDate, truncFunc, periods)
	if err != nil {
		return nil, fmt.Errorf("failed to query forecast data: %w", err)
	}
	defer rows.Close()

	var periods_data []models.PeriodStats

	for rows.Next() {
		var period models.PeriodStats
		var periodStart time.Time
		var income, expense int64

		if err := rows.Scan(&periodStart, &income, &expense); err != nil {
			return nil, fmt.Errorf("failed to scan forecast row: %w", err)
		}

		period.PeriodStart = periodStart
		period.PeriodEnd = calculatePeriodEnd(periodStart, groupBy)
		period.Income = income
		period.Expense = expense
		period.Balance = income - expense

		periods_data = append(periods_data, period)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating forecast rows: %w", err)
	}

	return periods_data, nil
}

func getTruncFunction(period models.TimePeriod) string {
	switch period {
	case models.TimePeriodMonth:
		return "month"
	case models.TimePeriodQuarter:
		return "quarter"
	case models.TimePeriodYear:
		return "year"
	default:
		return "month"
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

func (s *PostgresStorage) GetCategoryStatsByPeriods(ctx context.Context, userID string, startDate time.Time, periods int, groupBy models.TimePeriod) ([]models.CategoryPeriodStats, error) {
	truncFunc := getTruncFunction(groupBy)

	query := `
		WITH user_transactions AS (
			SELECT 
				t.mcc,
				t.amount,
				t.created_at
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			WHERE a.user_id = $1
				AND t.created_at >= $2
				AND t.type = 'EXPENSE'
		)
		SELECT 
			DATE_TRUNC($3, created_at) as period_start,
			COALESCE(mcc::TEXT, 'uncategorized') as category_id,
			SUM(amount) as total_amount
		FROM user_transactions
		GROUP BY DATE_TRUNC($3, created_at), mcc
		ORDER BY period_start DESC
		LIMIT $4 * 50
	`

	rows, err := s.pool.Query(ctx, query, userID, startDate, truncFunc, periods)
	if err != nil {
		return nil, fmt.Errorf("failed to query category stats: %w", err)
	}
	defer rows.Close()

	var stats []models.CategoryPeriodStats

	for rows.Next() {
		var stat models.CategoryPeriodStats
		if err := rows.Scan(&stat.PeriodStart, &stat.CategoryID, &stat.Amount); err != nil {
			return nil, fmt.Errorf("failed to scan category stat: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category stats: %w", err)
	}

	return stats, nil
}

func (s *PostgresStorage) GetRecurringPatterns(ctx context.Context, userID string) ([]models.RecurringPattern, error) {
	lookbackMonths := s.cfg.LookbackMonths
	minOccurrences := s.cfg.MinOccurrences
	intervalMinDays := s.cfg.IntervalMinDays
	intervalMaxDays := s.cfg.IntervalMaxDays

	query := fmt.Sprintf(`
		WITH user_expenses AS (
			SELECT 
				t.mcc,
				t.amount,
				t.created_at,
				LAG(t.created_at) OVER (
					PARTITION BY t.mcc
					ORDER BY t.created_at
				) as prev_date
			FROM transactions t
			JOIN accounts a ON t.account_id = a.id
			WHERE a.user_id = $1
				AND t.created_at >= NOW() - INTERVAL '%d months'
				AND t.type = 'EXPENSE'
				AND t.mcc IS NOT NULL
		)
		SELECT 
			mcc,
			PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY amount)::BIGINT as median_amount,
			AVG(EXTRACT(EPOCH FROM (created_at - prev_date))/86400) as avg_interval_days,
			MAX(created_at) as last_occurrence
		FROM user_expenses
		WHERE prev_date IS NOT NULL
		GROUP BY mcc
		HAVING COUNT(*) >= %d
			AND AVG(EXTRACT(EPOCH FROM (created_at - prev_date))/86400) BETWEEN %d AND %d
		ORDER BY last_occurrence DESC
	`, lookbackMonths, minOccurrences, intervalMinDays, intervalMaxDays)

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query recurring patterns: %w", err)
	}
	defer rows.Close()

	var patterns []models.RecurringPattern

	for rows.Next() {
		var pattern models.RecurringPattern
		if err := rows.Scan(&pattern.MCC, &pattern.MedianAmount, &pattern.AvgIntervalDays, &pattern.LastOccurrence); err != nil {
			return nil, fmt.Errorf("failed to scan recurring pattern: %w", err)
		}
		patterns = append(patterns, pattern)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recurring patterns: %w", err)
	}

	return patterns, nil
}
