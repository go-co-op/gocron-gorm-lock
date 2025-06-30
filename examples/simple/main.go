package main

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	gormlock "github.com/go-co-op/gocron-gorm-lock/v2"
)

func jobFunc() {
	fmt.Println("job func")
}

func main() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	worker := "example"
	err = db.AutoMigrate(&gormlock.CronJobLock{}) // We need the table to store the job execution
	if err != nil {
		panic(err)
	}

	locker, err := gormlock.NewGormLocker(db, worker)
	if err != nil {
		panic(err)
	}

	s, err := gocron.NewScheduler(gocron.WithDistributedLocker(locker))
	if err != nil {
		panic(err)
	}

	_, err = s.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(jobFunc), gocron.WithName("unique_name"))
	if err != nil {
		panic(err)
	}
	_, err = s.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(jobFunc), gocron.WithName("unique_name"))
	if err != nil {
		panic(err)
	}

	s.Start()
	time.Sleep(4 * time.Second)
}
