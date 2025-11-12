#!/bin/bash

HOST="localhost:50051"
USER_ID="11111111-1111-1111-1111-111111111111"

echo "Тестирование Analyzer API на $HOST"
echo "User ID: $USER_ID"
echo "=========================================="
echo ""

echo "1. GetStatistics - получение статистики за период (июнь-ноябрь 2025)"
echo "----------------------------------------------------------------------"
grpcurl -plaintext -d '{
  "user_id": "'$USER_ID'",
  "start_date": "2025-06-01T00:00:00Z",
  "end_date": "2025-11-30T23:59:59Z",
  "group_by": "TIME_PERIOD_MONTH"
}' $HOST analyzer.AnalyzerService/GetStatistics
echo ""
echo ""

echo "2. GetForecast - прогнозирование на 3 месяца вперед"
echo "----------------------------------------------------"
grpcurl -plaintext -d '{
  "user_id": "'$USER_ID'",
  "period": "TIME_PERIOD_MONTH",
  "periods_ahead": 3
}' $HOST analyzer.AnalyzerService/GetForecast
echo ""
echo ""

echo "3. GetAnomalies - детекция аномалий в тратах"
echo "----------------------------------------------"
grpcurl -plaintext -d '{
  "user_id": "'$USER_ID'",
  "period": "TIME_PERIOD_MONTH"
}' $HOST analyzer.AnalyzerService/GetAnomalies
echo ""
echo ""

echo "4. GetUpcomingRecurring - предсказание регулярных платежей"
echo "-----------------------------------------------------------"
grpcurl -plaintext -d '{
  "user_id": "'$USER_ID'"
}' $HOST analyzer.AnalyzerService/GetUpcomingRecurring
echo ""
echo ""

echo "=========================================="
echo "Тестирование завершено!"

