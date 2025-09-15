// Copyright 2025 James Ross
package obs

import (
    "strings"

    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func NewLogger(level string) (*zap.Logger, error) {
    lvl := zapcore.InfoLevel
    switch strings.ToLower(level) {
    case "debug":
        lvl = zapcore.DebugLevel
    case "warn":
        lvl = zapcore.WarnLevel
    case "error":
        lvl = zapcore.ErrorLevel
    }
    cfg := zap.NewProductionConfig()
    cfg.Level = zap.NewAtomicLevelAt(lvl)
    cfg.Encoding = "json"
    return cfg.Build()
}

// Convenience typed fields
func String(k, v string) zap.Field { return zap.String(k, v) }
func Int(k string, v int) zap.Field { return zap.Int(k, v) }
func Bool(k string, v bool) zap.Field { return zap.Bool(k, v) }
func Err(err error) zap.Field { return zap.Error(err) }
