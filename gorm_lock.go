package gormlock

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gorm.io/gorm"
)

func NewGormLocker(db *gorm.DB, worker string, options ...LockOption) (*GormLocker, error) {
	if db == nil {
		return nil, ErrGormCantBeNull
	}
	if worker == "" {
		return nil, ErrWorkerIsRequired
	}

	gl := &GormLocker{
		db:       db,
		worker:   worker,
		ttl:      defaultTTL,
		interval: defaultCleanInterval,
	}
	gl.jobIdentifier = defaultJobIdentifier(defaultPrecision)
	for _, option := range options {
		option(gl)
	}

	go func() {
		ticker := time.NewTicker(gl.interval)
		defer ticker.Stop()

		for range ticker.C {
			if gl.closed.Load() {
				return
			}

			gl.cleanExpiredRecords()
		}
	}()

	return gl, nil
}

var _ gocron.Locker = (*GormLocker)(nil)

type GormLocker struct {
	db            *gorm.DB
	worker        string
	ttl           time.Duration
	interval      time.Duration
	jobIdentifier func(ctx context.Context, key string) string

	closed atomic.Bool
}

func (g *GormLocker) cleanExpiredRecords() {
	g.db.Where("updated_at < ? and status = ?", time.Now().Add(-g.ttl), StatusFinished).Delete(&CronJobLock{})
}

func (g *GormLocker) Close() {
	g.closed.Store(true)
}

func (g *GormLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	ji := g.jobIdentifier(ctx, key)

	cjb := &CronJobLock{
		JobName:       key,
		JobIdentifier: ji,
		Worker:        g.worker,
		Status:        StatusRunning,
	}
	tx := g.db.WithContext(ctx).Create(cjb)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return &gormLock{db: g.db, id: cjb.GetID()}, nil
}

var _ gocron.Lock = (*gormLock)(nil)

type gormLock struct {
	db *gorm.DB
	//id the id that lock a particular job
	id int
}

func (g *gormLock) Unlock(_ context.Context) error {
	return g.db.Model(&CronJobLock{ID: g.id}).Updates(&CronJobLock{Status: StatusFinished}).Error
}
