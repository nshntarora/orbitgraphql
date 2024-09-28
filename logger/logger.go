package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

var logInstance *Logger

func init() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339
	logInstance = &Logger{
		logger: logger,
	}
}

func Error(ctx context.Context, message ...any) {
	logInstance.logger.Error().Fields(GetMetadata(ctx)).Msg(getMessage(message...))
}

func Info(ctx context.Context, message ...any) {
	logInstance.logger.Info().Fields(GetMetadata(ctx)).Msg(getMessage(message...))
}

func Debug(ctx context.Context, message ...any) {
	logInstance.logger.Debug().Fields(GetMetadata(ctx)).Msg(getMessage(message...))
}

func Warn(ctx context.Context, message ...any) {
	logInstance.logger.Warn().Fields(GetMetadata(ctx)).Msg(getMessage(message...))
}

func Fatal(ctx context.Context, message ...any) {
	logInstance.logger.Fatal().Fields(GetMetadata(ctx)).Msg(getMessage(message...))
}

func Panic(ctx context.Context, message ...any) {
	logInstance.logger.Panic().Fields(GetMetadata(ctx)).Msg(getMessage(message...))
}

func getMessage(message ...any) string {
	return fmt.Sprintf("%v", message...)
}

type loggerCtx struct{}

func SetMetadata(ctx context.Context, metadata map[string]interface{}) context.Context {
	existingMetadata := GetMetadata(ctx)
	if existingMetadata != nil {
		for k, v := range metadata {
			existingMetadata[k] = v
		}
		metadata = existingMetadata
	}
	return context.WithValue(ctx, loggerCtx{}, metadata)
}

func GetMetadata(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return nil
	}
	if metadata, ok := ctx.Value(loggerCtx{}).(map[string]interface{}); ok {
		return metadata
	}
	return nil
}
