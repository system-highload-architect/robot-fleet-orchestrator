package utils

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateID генерирует простой уникальный идентификатор на основе текущего времени и случайного числа.
// Для демонстрационных целей достаточно; в реальном проекте можно использовать snowflake или UUID.
func GenerateID(prefix string) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 6)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s-%d-%s", prefix, time.Now().UnixNano(), string(b))
}

// ContainsString проверяет, есть ли строка в слайсе.
func ContainsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Retry выполняет функцию fn до success или пока не кончатся попытки.
// Между попытками делает экспоненциальную задержку.
func Retry(attempts int, initialDelay time.Duration, fn func() error) error {
	delay := initialDelay
	for i := 0; i < attempts; i++ {
		err := fn()
		if err == nil {
			return nil
		}
		if i == attempts-1 {
			return err
		}
		time.Sleep(delay)
		delay *= 2
	}
	return nil
}

// Clamp ограничивает значение val между min и max.
func Clamp(val, min, max float64) float64 {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// Distance вычисляет евклидово расстояние между двумя точками (структура Point должна быть определена).
// Если Point определена в domain, мы не можем импортировать domain в pkg/utils, чтобы сохранить переносимость.
// Поэтому лучше оставить эту функцию в domain или в другом месте. Для демонстрации мы не включаем её сюда.
