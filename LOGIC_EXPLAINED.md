# Объяснение логики работы модуля аналитики

## Storage Layer (postgres.go)

### GetStatistics
Получает статистику по доходам/расходам за период с группировкой
- **Вход**: user_id, start_date, end_date, group_by (month/quarter/year)
- **Выход**: массив периодов с income, expense, balance, categories
- **SQL**: JOIN transactions+accounts, GROUP BY периодам, SUM агрегация

### getCategoryBreakdown  
Разбивка расходов по категориям (MCC кодам) за один период
- **Вход**: user_id, start_date, end_date
- **Выход**: массив {category_id, total_amount}
- **SQL**: GROUP BY mcc, ORDER BY amount DESC

### GetTransactionsForForecast
Получает последние N периодов для прогнозирования
- **Вход**: user_id, start_date, periods (сколько последних периодов взять)
- **Выход**: массив периодов, отсортированных DESC (новые → старые)
- **SQL**: аналогичен GetStatistics, но с LIMIT и ORDER BY DESC

### getTruncFunction
Преобразует models.TimePeriod → строку для SQL DATE_TRUNC
- MONTH → "month"
- QUARTER → "quarter"  
- YEAR → "year"

### calculatePeriodEnd
Вычисляет конец периода (последняя наносекунда)
- Для месяца: start + 1 month - 1 nanosecond
- Пример: 2024-01-01 → 2024-01-31 23:59:59.999999999

---

## Service Layer (analyzer.go)

### GetStatistics
Бизнес-логика получения статистики
- Валидация параметров (user_id, даты)
- Вызов storage.GetStatistics
- Подсчет totalIncome и totalExpense
- Логирование

### GetForecast
Бизнес-логика прогнозирования
- Валидация параметров
- Получение последних 6 периодов из БД
- Проверка достаточности данных (минимум 2 периода)
- Вызов calculateWMAForecast для расчета прогноза

### calculateWMAForecast
Weighted Moving Average - взвешенное скользящее среднее
- **Алгоритм**:
  1. Берем до 6 последних периодов
  2. Присваиваем веса: свежий период = 6, старый = 1
  3. Нормализуем веса (делим на сумму)
  4. Считаем avgIncome = Σ(income[i] * weight[i])
  5. Считаем avgExpense = Σ(expense[i] * weight[i])
  6. Генерируем N прогнозов с этими значениями


### Индексы (критичные):
```sql
CREATE INDEX idx_transactions_account_date_type 
  ON transactions(account_id, created_at, type);

CREATE INDEX idx_accounts_user_id 
  ON accounts(user_id);

CREATE INDEX idx_transactions_mcc 
  ON transactions(mcc);
```


