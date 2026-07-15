#!/bin/bash
set -e

# Список директорий, где нужен go.mod (относительно корня проекта)
MODULES=(
    "cmd/orchestrator"
    "internal/modules/analytics"
    "internal/modules/robot-manager"
    "internal/modules/task-dispatcher"
    "internal/modules/location-service"
    "pkg/logger"
    "pkg/config"
    "pkg/utils"
    "emulator/cmd/emulator"
)

echo "=== Создание go.mod для всех модулей ==="

for mod in "${MODULES[@]}"; do
    if [ ! -f "$mod/go.mod" ]; then
        echo "Создаю go.mod для $mod"
        module_name="robot-fleet-orchestrator/$mod"
        (cd "$mod" && go mod init "$module_name")
    else
        echo "go.mod уже существует в $mod"
    fi
done

echo ""
echo "=== Настройка go.work ==="

if [ ! -f "go.work" ]; then
    echo "Создаю go.work"
    go work init
fi

# Добавляем все модули в go.work (если ещё не добавлены)
for mod in "${MODULES[@]}"; do
    echo "Добавляю $mod в go.work"
    go work use "$mod"
done

echo ""
echo "=== Синхронизация go.work ==="
go work sync

echo ""
echo "✅ Готово! Все модули инициализированы и синхронизированы."