# Развертывание

## Локальная разработка
```bash
make run          # запуск через `go run`
docker-compose up # поднять PostgreSQL, Redis
```

## Docker-образ
```bash
docker build -t robot-orchestrator .
docker run -p 8080:8080 robot-orchestrator
```

## Kubernetes (этап 2+)
- Манифесты лежат в `deploy/k8s/`.
- Использовать ConfigMap для переменных окружения.
- HPA для горизонтального масштабирования.
