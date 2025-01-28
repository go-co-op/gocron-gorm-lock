package gormlock

import (
	"time"

	"gorm.io/gorm"
)

type JobLock[T any] interface {
	GetID() T
	SetJobIdentifier(ji string)
}

var _ JobLock[int] = (*CronJobLock)(nil)

type CronJobLock struct {
	ID            int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	JobName       string `gorm:"index:idx_name,unique"`
	JobIdentifier string `gorm:"index:idx_name,unique"`
	Worker        string `gorm:"not null"`
	Status        string `gorm:"not null"`
}

func (cjb *CronJobLock) SetJobIdentifier(ji string) {
	cjb.JobIdentifier = ji
}

func (cjb *CronJobLock) GetID() int {
	return cjb.ID
}

func (cjb *CronJobLock) BeforeCreate(_ *gorm.DB) error {
	if cjb.Worker == "" {
		return ErrWorkerIsRequired
	}
	if cjb.Status == "" {
		return ErrStatusIsRequired
	}
	return nil
}
