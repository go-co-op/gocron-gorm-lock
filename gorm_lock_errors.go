package gormlock

import "errors"

var (
	ErrGormCantBeNull   = errors.New("gorm can't be null")
	ErrWorkerIsRequired = errors.New("worker is required")
	ErrStatusIsRequired = errors.New("status is required")
)
