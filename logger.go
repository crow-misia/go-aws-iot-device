// Copyright 2024 Zenichi Amano.

package awsiotdevice

import (
	"fmt"
	"log/slog"

	"golang.org/x/net/context"
)

type SlogLogger struct {
	ctx    context.Context
	logger *slog.Logger
	level  slog.Level
}

func NewSlogLogger(ctx context.Context, logger *slog.Logger, level slog.Level) *SlogLogger {
	return &SlogLogger{
		ctx:    ctx,
		logger: logger,
		level:  level,
	}
}

func (s SlogLogger) Println(v ...any) {
	s.logger.Log(s.ctx, s.level, fmt.Sprintf("%v", v...))
}

func (s SlogLogger) Printf(format string, v ...any) {
	s.logger.Log(s.ctx, s.level, fmt.Sprintf(format, v...))
}
