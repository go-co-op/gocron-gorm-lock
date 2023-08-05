package gormlock

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"
)

func NewGormLocker(db *gorm.DB, worker string) (gocron.Locker, error) {
	return &gormLocker{db: db, worker: worker}, nil
}

var _ gocron.Locker = (*gormLocker)(nil)

type gormLocker struct {
	db     *gorm.DB
	worker string
}

func (g *gormLocker) Lock(_ context.Context, key string) (gocron.Lock, error) {
	// This is a complicated thing, we are assuming that the minute precision is enough
	ji := time.Now().Truncate(time.Millisecond).Format("2006-01-02 15:04:05.000")

	// I would like that people can "pass" their own implementation,
	cjb := &CronJobLock{
		JobName:       key,
		JobIdentifier: ji,
		Worker:        g.worker,
		Status:        "RUNNING",
	}
	tx := g.db.Create(cjb)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &gormLock{db: g.db, id: cjb.GetId()}, nil
}

var _ gocron.Lock = (*gormLock)(nil)

type gormLock struct {
	db *gorm.DB
	id int
}

func (g *gormLock) Unlock(_ context.Context) error {
	return g.db.Model(&CronJobLock{Id: g.id}).Updates(&CronJobLock{Status: "FINISHED"}).Error
}
