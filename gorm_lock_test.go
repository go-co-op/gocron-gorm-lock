package gormlock

import (
	"context"
	"testing"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	testcontainerspostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestNewGormLocker_Validation(t *testing.T) {
	tests := map[string]struct {
		db     *gorm.DB
		worker string
		err    string
	}{
		"db is nil":       {db: nil, worker: "local", err: "gorm db definition can't be null"},
		"worker is empty": {db: &gorm.DB{}, worker: "", err: "worker name can't be null"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := NewGormLocker(tc.db, tc.worker)
			if assert.Error(t, err) {
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}

func TestEnableDistributedLocking(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := testcontainerspostgres.RunContainer(ctx,
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
	s1 := gocron.NewScheduler(time.UTC)
	s1.WithDistributedLocker(l1)
	_, err = s1.Every("1s").Do(f, 1)
	require.NoError(t, err)

	l2, err := NewGormLocker(db, "s2")
	require.NoError(t, err)
	s2 := gocron.NewScheduler(time.UTC)
	s2.WithDistributedLocker(l2)
	_, err = s2.Every("1s").Do(f, 2)
	require.NoError(t, err)

	s1.StartAsync()
	s2.StartAsync()

	time.Sleep(3500 * time.Millisecond)

	s1.Stop()
	s2.Stop()
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
	postgresContainer, err := testcontainerspostgres.RunContainer(ctx,
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

	s1 := gocron.NewScheduler(time.UTC)
	l1, err := NewGormLocker(db, "s1")
	require.NoError(t, err)
	s1.WithDistributedLocker(l1)
	_, err = s1.Every("1s").Name("f").Do(f, 1)
	require.NoError(t, err)
	_, err = s1.Every("1s").Name("f2").Do(f2, 1)
	require.NoError(t, err)

	s2 := gocron.NewScheduler(time.UTC)
	l2, err := NewGormLocker(db, "s2")
	require.NoError(t, err)
	s2.WithDistributedLocker(l2)
	_, err = s2.Every("1s").Name("f").Do(f, 2)
	require.NoError(t, err)
	_, err = s2.Every("1s").Name("f2").Do(f2, 2)
	require.NoError(t, err)

	s3 := gocron.NewScheduler(time.UTC)
	l3, err := NewGormLocker(db, "s3")
	require.NoError(t, err)
	s3.WithDistributedLocker(l3)
	_, err = s3.Every("1s").Name("f").Do(f, 3)
	require.NoError(t, err)
	_, err = s3.Every("1s").Name("f2").Do(f2, 3)
	require.NoError(t, err)

	s1.StartAsync()
	s2.StartAsync()
	s3.StartAsync()

	time.Sleep(3500 * time.Millisecond)

	s1.Stop()
	s2.Stop()
	s3.Stop()
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
	ctx := context.Background()
	postgresContainer, err := testcontainerspostgres.RunContainer(ctx,
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

	precision := 60 * time.Minute
	// creating an entry to force the unique identifier error
	cjb := &CronJobLock{
		JobName:       "job",
		JobIdentifier: time.Now().Truncate(precision).Format("2006-01-02 15:04:05.000"),
		Worker:        "local",
		Status:        StatusRunning,
	}
	require.NoError(t, db.Create(cjb).Error)

	l, _ := NewGormLocker(db, "local", WithDefaultJobIdentifier(precision))
	_, lerr := l.Lock(ctx, "job")
	if assert.Error(t, lerr) {
		assert.ErrorContains(t, lerr, "violates unique constraint")
	}
}

func TestHandleTTL(t *testing.T) {
	ctx := context.Background()
	postgresContainer, err := testcontainerspostgres.RunContainer(ctx,
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

	s1 := gocron.NewScheduler(time.UTC)
	s1.WithDistributedLocker(l1)

	_, err = s1.Every("1s").Do(func() {})
	require.NoError(t, err)

	s1.StartAsync()

	time.Sleep(3500 * time.Millisecond)

	s1.Stop()

	var allCronJobs []*CronJobLock
	db.Find(&allCronJobs)
	assert.GreaterOrEqual(t, len(allCronJobs), 3)

	// wait for data to expire
	time.Sleep(1500 * time.Millisecond)
	l1.cleanExpire()
	db.Find(&allCronJobs)
	assert.Equal(t, 0, len(allCronJobs))
}
