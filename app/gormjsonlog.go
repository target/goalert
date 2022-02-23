package app

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/target/goalert/util/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type gormJSONLogger struct{}

func (l *gormJSONLogger) LogMode(logger.LogLevel) logger.Interface { return l }

func (l *gormJSONLogger) Info(ctx context.Context, s string, args ...interface{}) {
	log.Logf(ctx, s, args...)
}

func (l *gormJSONLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	log.Logf(ctx, s, args...)
}

func (l *gormJSONLogger) Error(ctx context.Context, s string, args ...interface{}) {
	log.Log(ctx, fmt.Errorf(s, args...))
}

func (l *gormJSONLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err == nil || errors.Is(err, gorm.ErrRecordNotFound) {
		return
	}

	elapsed := time.Since(begin)
	sql, rowCount := fc()
	fields := log.Fields{
		"ElapsedMs": elapsed.Seconds() * 1000,
		"Source":    utils.FileWithLineNum(),
		"Rows":      rowCount,
		"SQL":       sql,
	}

	ctx = log.WithFields(ctx, fields)
	if err != nil {
		log.Log(ctx, err)
	} else {
		log.Logf(ctx, "SQL Trace.")
	}
}
