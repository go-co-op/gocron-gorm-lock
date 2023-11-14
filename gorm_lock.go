package gormlock

import (
	"context"
	"time"

	"github.com/go-co-op/gocron"
	"gorm.io/gorm"
)

var (
	defaultPrecision = time.Second

	StatusRunning  = "RUNNING"
	StatusFinished = "FINISHED"
)

func NewGormLocker(db *gorm.DB, worker string, options ...LockOption) (gocron.Locker, error) {
	gl := &gormLocker{db: db, worker: worker}
	for _, option := range options {
		option(gl)
	}
	return gl, nil
}

var _ gocron.Locker = (*gormLocker)(nil)

type gormLocker struct {
	db            *gorm.DB
	worker        string
	jobIdentifier func(ctx context.Context, key string) string
}

func (g *gormLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	ji := g.getJobIdentifier(ctx, key)

	// I would like that people can "pass" their own implementation,
	cjb := &CronJobLock{
		JobName:       key,
		JobIdentifier: ji,
		Worker:        g.worker,
		Status:        StatusRunning,
	}
	tx := g.db.Create(cjb)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &gormLock{db: g.db, id: cjb.GetID()}, nil
}

func (g *gormLocker) getJobIdentifier(ctx context.Context, key string) string {
	if g.jobIdentifier == nil {
		return time.Now().Truncate(defaultPrecision).Format("2006-01-02 15:04:05.000")
	}
	return g.jobIdentifier(ctx, key)
}

var _ gocron.Lock = (*gormLock)(nil)

type gormLock struct {
	db *gorm.DB
	id int
}

func (g *gormLock) Unlock(_ context.Context) error {
	return g.db.Model(&CronJobLock{ID: g.id}).Updates(&CronJobLock{Status: StatusFinished}).Error
}
