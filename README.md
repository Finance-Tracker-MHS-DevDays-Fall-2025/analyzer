# analyzer

Микросервис-анализатор на Go с использованием gRPC.

## Быстрый старт



### 1. Установить зависимости

```bash
go mod download
```

### 2. Сгенерировать код из protobuf

```bash
make proto
```

### 3. Запустить сервис

```bash
make run
```

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
