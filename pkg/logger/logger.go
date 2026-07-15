package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

// Logger - обёртка над slog.Logger с методами уровней.
type Logger struct {
	logger *slog.Logger
}

// Option задаёт конфигурацию логгера.
type Option func(*options)

type options struct {
	level      slog.Level
	output     io.Writer
	addSource  bool
	jsonFormat bool
}

// WithLevel устанавливает минимальный уровень логирования.
func WithLevel(level slog.Level) Option {
	return func(o *options) { o.level = level }
}

// WithOutput устанавливает вывод (по умолчанию os.Stdout).
func WithOutput(w io.Writer) Option {
	return func(o *options) { o.output = w }
}

// WithSource добавляет информацию о файле и строке в записи.
func WithSource(add bool) Option {
	return func(o *options) { o.addSource = add }
}

// WithJSONFormat включает JSON-формат (по умолчанию текстовый для удобства чтения).
func WithJSONFormat(json bool) Option {
	return func(o *options) { o.jsonFormat = json }
}

// New создаёт новый логгер с опциями.
func New(opts ...Option) *Logger {
	o := &options{
		level:      slog.LevelInfo,
		output:     os.Stdout,
		addSource:  false,
		jsonFormat: false,
	}
	for _, opt := range opts {
		opt(o)
	}

	var handler slog.Handler
	if o.jsonFormat {
		handler = slog.NewJSONHandler(o.output, &slog.HandlerOptions{
			Level:     o.level,
			AddSource: o.addSource,
		})
	} else {
		handler = slog.NewTextHandler(o.output, &slog.HandlerOptions{
			Level:     o.level,
			AddSource: o.addSource,
		})
	}
	return &Logger{
		logger: slog.New(handler),
	}
}

// Default возвращает логгер с настройками по умолчанию (уровень Info, текстовый вывод).
func Default() *Logger {
	return New()
}

// With создаёт дочерний логгер с привязанными полями.
func (l *Logger) With(fields ...interface{}) *Logger {
	if len(fields)%2 != 0 {
		// Нечётное количество полей — игнорируем последний
		fields = fields[:len(fields)-1]
	}
	args := make([]any, 0, len(fields))
	for i := 0; i < len(fields); i += 2 {
		args = append(args, slog.Any(fields[i].(string), fields[i+1]))
	}
	return &Logger{
		logger: l.logger.With(args...),
	}
}

// WithContext создаёт логгер с полями из context (если нужно, можно расширить).
// В данном примере просто возвращает тот же логгер.
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Можно извлечь поля из контекста (например, requestID) и добавить.
	// Для демонстрации оставляем как есть.
	return l
}

// Debug выводит сообщение уровня Debug.
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.logger.Debug(msg, args...)
}

// Info выводит сообщение уровня Info.
func (l *Logger) Info(msg string, args ...interface{}) {
	l.logger.Info(msg, args...)
}

// Warn выводит сообщение уровня Warn.
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.logger.Warn(msg, args...)
}

// Error выводит сообщение уровня Error.
func (l *Logger) Error(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
}

// Fatal выводит сообщение уровня Error и завершает программу.
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.logger.Error(msg, args...)
	os.Exit(1)
}

// Sync вызывается для сброса буферов (если нужно).
func (l *Logger) Sync() error {
	// slog не имеет Sync, но мы оставим для совместимости.
	return nil
}
