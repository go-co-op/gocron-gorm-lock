package gormlock

import "time"

type JobLock[T any] interface {
	GetId() T
	SetJobIdentifier(ji string)
}

var _ JobLock[int] = (*CronJobLock)(nil)

type CronJobLock struct {
	Id            int
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

func (cjb *CronJobLock) GetId() int {
	return cjb.Id
}
