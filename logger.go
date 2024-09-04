// Copyright 2024 Zenichi Amano.

package awsiotdevice

import (
	"fmt"
	"golang.org/x/net/context"
	"log/slog"
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

func (s SlogLogger) Println(v ...interface{}) {
	s.logger.Log(s.ctx, s.level, fmt.Sprintf("%v", v...))
}

func (s SlogLogger) Printf(format string, v ...interface{}) {
	s.logger.Log(s.ctx, s.level, fmt.Sprintf(format, v...))
}
