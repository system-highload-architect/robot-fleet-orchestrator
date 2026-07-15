# Запуск проекта

## Предварительные требования
- Go 1.23 или новее (рекомендуется 1.25+ для полной совместимости)
- (Опционально) Docker и Docker Compose для поднятия PostgreSQL/Redis — пока не используются, но структура готова
- (Опционально) `golangci-lint` для проверки кода (автоматически устанавливается через `make install-lint`)

## 1. Клонирование репозитория
```bash
git clone <repository-url>
cd robot-fleet-orchestrator
```

## 2. Инициализация модулей и синхронизация
Проект использует `go.work` для локальной разработки нескольких модулей.

```bash
# Автоматическое создание go.mod для всех модулей (рекомендуется)
./scripts/setup-go-modules.sh

# Или вручную:
go work init
go work use ./cmd/orchestrator
go work use ./internal/modules/robot-manager
go work use ./internal/modules/task-dispatcher
go work use ./internal/modules/location-service
go work use ./internal/modules/analytics
go work use ./pkg/logger
go work use ./pkg/config
go work use ./pkg/utils
go work use ./emulator/cmd/emulator

# Синхронизация зависимостей
go work sync
```

## 3. Сборка и запуск оркестратора
```bash
# Установка зависимостей (если ещё не сделано)
go mod download

# Запуск оркестратора (HTTP-сервер на порту 8080)
go run cmd/orchestrator/main.go
```

### Демонстрационный режим (оркестратор + эмулятор)
```bash
make demo
```

Или используйте Makefile (удобнее):
```bash
make run          # только оркестратор (в фоне или терминале)
make dev          # быстрый запуск оркестратора без сборки
```

## 4. Запуск эмулятора (в отдельном терминале)
Эмулятор имитирует работу роботов и создаёт задачи. Он общается с оркестратором через HTTP API.

```bash
# Настройка переменных окружения (опционально)
export ORCHESTRATOR_URL="http://localhost:8080"  # по умолчанию
export ROBOT_COUNT=5                              # количество роботов
export TASK_INTERVAL_SEC=10                       # интервал создания задач

# Запуск эмулятора
go run emulator/cmd/emulator/main.go
```

## 5. Проверка работы
Откройте в браузере или выполните через `curl`:
- `GET http://localhost:8080/api/v1/robots` — список роботов.
- `GET http://localhost:8080/api/v1/tasks` — список задач.
- `GET http://localhost:8080/api/v1/analytics/robot-stats` — статистика.

В логах оркестратора и эмулятора будет видно взаимодействие.

## 6. Остановка
Нажмите `Ctrl+C` в обоих терминалах.

## 7. Тестирование и линтинг
```bash
# Запуск тестов
make test

# Проверка кода линтером (автоустановка, если отсутствует)
make lint

# Полная проверка (линтер + сборка)
make check
```

## 8. Дополнительные команды (без Makefile)
```bash
# Запуск тестов напрямую
go test ./... -v

# Линтер (если установлен golangci-lint)
golangci-lint run ./...
```

## Структура проекта
Подробнее см. в [README.md](README.md) и [архитектурном документе](docs/architecture.md).