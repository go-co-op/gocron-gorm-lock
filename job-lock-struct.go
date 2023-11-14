package gormlock

import "time"

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
	Worker        string `gorm:"not null;default:null"`
	Status        string `gorm:"not nullldefault:null"`
}

func (cjb *CronJobLock) SetJobIdentifier(ji string) {
	cjb.JobIdentifier = ji
}

func (cjb *CronJobLock) GetID() int {
	return cjb.ID
}
