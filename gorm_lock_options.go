package gormlock

import (
	"context"
	"time"
)

var (
	defaultPrecision     = time.Second
	defaultJobIdentifier = func(precision time.Duration) func(ctx context.Context, key string) string {
		return func(_ context.Context, _ string) string {
			return time.Now().Truncate(precision).Format("2006-01-02 15:04:05.000")
		}
	}

	StatusRunning  = "RUNNING"
	StatusFinished = "FINISHED"

	defaultTTL           = 24 * time.Hour
	defaultCleanInterval = 5 * time.Second
)

type LockOption func(*GormLocker)

// WithJobIdentifier overrides the default job identifier with your own custom implementation
func WithJobIdentifier(f func(ctx context.Context, key string) string) LockOption {
	return func(l *GormLocker) {
		l.jobIdentifier = f
	}
}

func WithDefaultJobIdentifier(precision time.Duration) LockOption {
	return func(l *GormLocker) {
		l.jobIdentifier = defaultJobIdentifier(precision)
	}
}

// WithTTL when the locker records in the database exceeds the ttl, it is cleaned up.
// to avoid excessive data in the database.
func WithTTL(ttl time.Duration) LockOption {
	return func(l *GormLocker) {
		l.ttl = ttl
	}
}

// WithCleanInterval the time interval to run clean operation.
func WithCleanInterval(interval time.Duration) LockOption {
	return func(l *GormLocker) {
		l.interval = interval
	}
}
