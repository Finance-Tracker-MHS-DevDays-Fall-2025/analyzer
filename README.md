# Analyzer Service

Микросервис аналитики для финансового трекера. Предоставляет статистику по доходам/расходам и прогнозирование с использованием WMA (Weighted Moving Average).

## Функциональность

- **GetStatistics** - статистика по доходам/расходам с группировкой по периодам и категориям
- **GetForecast** - прогноз на N периодов вперед на основе исторических данных

## Быстрый старт

### 1. Настроить переменные окружения

Создайте файл `.env` в корне проекта:

```bash
# .env
DB_HOST=
DB_PORT=6432
DB_USER=service
DB_PASSWORD=your_password_here
DB_NAME=master
DB_SSLMODE=require
```

### 2. Установить зависимости

```bash
go mod download
```

### 3. Сгенерировать код из protobuf

```bash
make proto
```

### 4. Запустить тесты

```bash
go test ./internal/... -v
```

### 5. Запустить сервис

```bash
make run
```

Сервис будет доступен на:

- gRPC: `localhost:50051`
- Debug HTTP: `localhost:8080`

## Команды

- `make init-submodule` - инструкция по добавлению submodule
- `make update-submodule` - обновить proto файлы из submodule
- `make proto` - сгенерировать Go код из protobuf
- `make build` - собрать бинарник
- `make run` - запустить сервис
- `make clean` - очистить сгенерированные файлы

## Структура проекта

```
analyzer/
├── api/              # git submodule с protobuf файлами
├── cmd/
│   └── analyzer/     # точка входа приложения
├── pkg/
│   └── api/          # сгенерированный код из proto (не коммитится)
├── go.mod
├── Makefile
└── README.md
```
