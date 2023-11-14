package gormlock

import (
	"context"
)

type LockOption func(*gormLocker)

func WithJobIdentifier(f func(ctx context.Context, key string) string) LockOption {
	return func(l *gormLocker) {
		l.jobIdentifier = f
	}
}
