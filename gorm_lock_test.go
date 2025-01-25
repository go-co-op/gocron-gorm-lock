package gormlock

import (
	"context"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewGormLocker_Validation(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		db     *gorm.DB
		worker string
		err    string
	}{
		"db is nil":       {db: nil, worker: "local", err: ErrGormCantBeNull.Error()},
		"worker is empty": {db: &gorm.DB{}, worker: "", err: ErrWorkerIsRequired.Error()},
	}

	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := NewGormLocker(tc.db, tc.worker)
			if assert.Error(t, err) {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestEnableDistributedLocking(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := testcontainerspostgres.Run(ctx, "docker.io/postgres:16-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(5*time.Second)))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable", "application_name=test")
	assert.NoError(t, err)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&CronJobLock{})
	require.NoError(t, err)

	resultChan := make(chan int, 10)
	f := func(schedulerInstance int) {
		resultChan <- schedulerInstance
		println(time.Now().Truncate(defaultPrecision).Format("2006-01-02 15:04:05.000"))
	}

	l1, err := NewGormLocker(db, "s1")
	require.NoError(t, err)
	s1, schErr := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithDistributedLocker(l1))
	require.NoError(t, schErr)

	_, err = s1.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f, 1))
	require.NoError(t, err)

	l2, err := NewGormLocker(db, "s2")
	require.NoError(t, err)
	s2, schErr := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithDistributedLocker(l2))
	require.NoError(t, schErr)
	_, err = s2.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f, 2))
	require.NoError(t, err)

	s1.Start()
	s2.Start()

	time.Sleep(4500 * time.Millisecond)

	require.NoError(t, s1.Shutdown())
	require.NoError(t, s2.Shutdown())
	close(resultChan)

	var results []int
	for r := range resultChan {
		results = append(results, r)
	}
	assert.Len(t, results, 4)
	var allCronJobs []*CronJobLock
	db.Find(&allCronJobs)
	assert.Equal(t, len(results), len(allCronJobs))
}

func TestEnableDistributedLocking_DifferentJob(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := testcontainerspostgres.Run(ctx, "docker.io/postgres:16-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(5*time.Second)))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable", "application_name=test")
	assert.NoError(t, err)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&CronJobLock{})
	require.NoError(t, err)

	resultChan := make(chan int, 10)
	f := func(schedulerInstance int) {
		resultChan <- schedulerInstance
	}

	result2Chan := make(chan int, 10)
	f2 := func(schedulerInstance int) {
		result2Chan <- schedulerInstance
	}

	l1, err := NewGormLocker(db, "s1")
	require.NoError(t, err)
	s1, schErr := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithDistributedLocker(l1))
	require.NoError(t, schErr)
	_, err = s1.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f, 1), gocron.WithName("f"))
	require.NoError(t, err)
	_, err = s1.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f2, 1), gocron.WithName("f2"))
	require.NoError(t, err)

	l2, err := NewGormLocker(db, "s2")
	require.NoError(t, err)
	s2, schErr := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithDistributedLocker(l2))
	require.NoError(t, schErr)
	_, err = s2.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f, 2), gocron.WithName("f"))
	require.NoError(t, err)
	_, err = s2.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f2, 2), gocron.WithName("f2"))
	require.NoError(t, err)

	l3, err := NewGormLocker(db, "s3")
	require.NoError(t, err)
	s3, schErr := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithDistributedLocker(l3))
	require.NoError(t, schErr)

	_, err = s3.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f, 3), gocron.WithName("f"))
	require.NoError(t, err)
	_, err = s3.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f2, 3), gocron.WithName("f2"))
	require.NoError(t, err)

	s1.Start()
	s2.Start()
	s3.Start()

	time.Sleep(4500 * time.Millisecond)

	require.NoError(t, s1.Shutdown())
	require.NoError(t, s2.Shutdown())
	require.NoError(t, s3.Shutdown())
	close(resultChan)
	close(result2Chan)

	var results []int
	for r := range resultChan {
		results = append(results, r)
	}
	assert.Len(t, results, 4, "f is expected 4 times")
	var results2 []int
	for r := range result2Chan {
		results2 = append(results2, r)
	}
	assert.Len(t, results2, 4, "f2 is expected 4 times")
	var allCronJobs []*CronJobLock
	db.Find(&allCronJobs)
	assert.Equal(t, len(results)+len(results2), len(allCronJobs))
}

func TestJobReturningExceptionWhenUnique(t *testing.T) {
	precision := 60 * time.Minute
	tests := map[string]struct {
		ji         string
		lockOption []LockOption
	}{
		"default job identifier": {
			ji:         defaultJobIdentifier(precision)(context.Background(), "key"),
			lockOption: []LockOption{WithDefaultJobIdentifier(precision)},
		},
		"override job identifier with hardcoded name": {
			ji: "hardcoded",
			lockOption: []LockOption{WithJobIdentifier(func(_ context.Context, _ string) string {
				return "hardcoded"
			})},
		},
	}
	for name, tc := range tests {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			postgresContainer, err := testcontainerspostgres.Run(ctx, "docker.io/postgres:16-alpine",
				testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).WithStartupTimeout(5*time.Second)))
			require.NoError(t, err)
			t.Cleanup(func() {
				if err := postgresContainer.Terminate(ctx); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			})

			connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable", "application_name=test")
			assert.NoError(t, err)

			db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
			require.NoError(t, err)

			err = db.AutoMigrate(&CronJobLock{})
			require.NoError(t, err)

			// creating an entry to force the unique identifier error
			cjb := &CronJobLock{
				JobName:       "job",
				JobIdentifier: tc.ji,
				Worker:        "local",
				Status:        StatusRunning,
			}
			require.NoError(t, db.Create(cjb).Error)

			l, _ := NewGormLocker(db, "local", tc.lockOption...)
			_, lerr := l.Lock(ctx, "job")
			if assert.Error(t, lerr) {
				assert.ErrorContains(t, lerr, "violates unique constraint")
			}
		})
	}
}

func TestHandleTTL(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := testcontainerspostgres.Run(ctx, "docker.io/postgres:16-alpine",
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(5*time.Second)))
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	})

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable", "application_name=test")
	assert.NoError(t, err)

	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&CronJobLock{})
	require.NoError(t, err)

	l1, err := NewGormLocker(db, "s1", WithTTL(1*time.Second))
	require.NoError(t, err)

	s1, schErr := gocron.NewScheduler(gocron.WithLocation(time.UTC), gocron.WithDistributedLocker(l1))
	require.NoError(t, schErr)

	_, err = s1.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(func() {}))
	require.NoError(t, err)

	s1.Start()

	time.Sleep(3500 * time.Millisecond)

	require.NoError(t, s1.Shutdown())

	var allCronJobs []*CronJobLock
	db.Find(&allCronJobs)
	assert.GreaterOrEqual(t, len(allCronJobs), 3)

	// wait for data to expire
	time.Sleep(1500 * time.Millisecond)
	l1.cleanExpiredRecords()
	db.Find(&allCronJobs)
	assert.Equal(t, 0, len(allCronJobs))
}
