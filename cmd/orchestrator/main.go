package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"robot-fleet-orchestrator/internal/infrastructure/eventbus"
	"robot-fleet-orchestrator/internal/modules/analytics/api"
	analyticsRepo "robot-fleet-orchestrator/internal/modules/analytics/repository"
	analyticsSvc "robot-fleet-orchestrator/internal/modules/analytics/service"
	locationAPI "robot-fleet-orchestrator/internal/modules/location-service/api"
	locationRepo "robot-fleet-orchestrator/internal/modules/location-service/repository"
	locationSvc "robot-fleet-orchestrator/internal/modules/location-service/service"
	robotAPI "robot-fleet-orchestrator/internal/modules/robot-manager/api"
	robotRepo "robot-fleet-orchestrator/internal/modules/robot-manager/repository"
	robotSvc "robot-fleet-orchestrator/internal/modules/robot-manager/service"
	taskAPI "robot-fleet-orchestrator/internal/modules/task-dispatcher/api"
	taskRepo "robot-fleet-orchestrator/internal/modules/task-dispatcher/repository"
	taskSvc "robot-fleet-orchestrator/internal/modules/task-dispatcher/service"
	"robot-fleet-orchestrator/internal/shared/domain"
	"robot-fleet-orchestrator/pkg/config"
	"robot-fleet-orchestrator/pkg/logger"
)

func main() {
	logger := logger.Default()

	// Загружаем конфигурацию
	var cfg struct {
		HTTPPort int `yaml:"http_port" env:"HTTP_PORT"`
	}
	if err := config.Load(&cfg, config.WithPath("config.yaml"), config.WithEnvPrefix("ROBOT_")); err != nil {
		logger.Warn("Config load failed, using defaults", "error", err)
		cfg.HTTPPort = 8080
	}
	// Если порт всё ещё 0, ставим дефолт
	if cfg.HTTPPort == 0 {
		cfg.HTTPPort = 8080
	}

	// Инициализируем шину событий (in-memory)
	bus := eventbus.NewInMemoryBus()

	// === Инициализация репозиториев ===
	robotRepo := robotRepo.NewInMemoryRepository()
	taskRepo := taskRepo.NewInMemoryRepository()
	locationRepo := locationRepo.NewInMemoryRepository()
	analyticsRepo := analyticsRepo.NewInMemoryRepository()

	// === Инициализация сервисов ===
	robotSvc := robotSvc.NewService(robotRepo, bus, logger)
	taskSvc := taskSvc.NewService(taskRepo, bus, logger) // Убрали лишний аргумент
	locationSvc := locationSvc.NewService(locationRepo, bus, logger)
	analyticsSvc := analyticsSvc.NewService(analyticsRepo, bus, logger)

	// === Инициализация HTTP-обработчиков ===
	robotHandler := robotAPI.NewHandler(robotSvc, logger)
	taskHandler := taskAPI.NewHandler(taskSvc, logger)
	locationHandler := locationAPI.NewHandler(locationSvc, logger)
	analyticsHandler := api.NewHandler(analyticsSvc, logger)

	// === Регистрация маршрутов ===
	mux := http.NewServeMux()
	robotHandler.RegisterRoutes(mux)
	taskHandler.RegisterRoutes(mux)
	locationHandler.RegisterRoutes(mux)
	analyticsHandler.RegisterRoutes(mux)

	// === Создание тестовых данных (для демонстрации) ===
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Добавляем роботов напрямую через репозиторий, чтобы не зависеть от метода в сервисе
	robotRepo.Save(ctx, &domain.Robot{ID: "robot-1", Type: domain.TypeDelivery, Status: domain.StatusIdle})
	robotRepo.Save(ctx, &domain.Robot{ID: "robot-2", Type: domain.TypeDrone, Status: domain.StatusIdle})
	robotRepo.Save(ctx, &domain.Robot{ID: "robot-3", Type: domain.TypeLoader, Status: domain.StatusIdle})

	// Добавляем локацию
	locationSvc.CreateLocation(ctx, "Main Warehouse", "warehouse")

	logger.Info("Orchestrator started", "port", cfg.HTTPPort)

	// === Запуск HTTP-сервера ===
	addr := ":" + strconv.Itoa(cfg.HTTPPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP server failed", "error", err)
		}
	}()

	// === Ожидание сигнала завершения ===
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info("Shutting down gracefully...")

	// Останавливаем сервисы
	analyticsSvc.Shutdown()

	// Закрываем HTTP-сервер
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error("HTTP server shutdown error", "error", err)
	}

	// Закрываем шину
	if err := bus.Close(); err != nil {
		logger.Error("Event bus close error", "error", err)
	}

	logger.Info("Orchestrator stopped")
}
