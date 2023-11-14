package gormlock

import (
	"context"
	"time"
)

type LockOption func(*gormLocker)

func WithJobIdentifier(f func(ctx context.Context, key string) string) LockOption {
	return func(l *gormLocker) {
		l.jobIdentifier = f
	}
}

func WithDefaultJobIdentifier(precision time.Duration) LockOption {
	return func(l *gormLocker) {
		l.jobIdentifier = defaultJobIdentifier(precision)
	}
}
