# gocron-gorm-lock
A gocron locker implementation using gorm

## install

```
go get github.com/go-co-op/gocron-gorm-lock
```

## usage

Here is an example usage that would be deployed in multiple instances

```go
package main

import (
	"fmt"

	"github.com/go-co-op/gocron"
	gormlock "github.com/go-co-op/gocron-gorm-lock"
	"gorm.io/gorm"
	"time"
)

func main() {
	var db * gorm.DB // gorm db connection
	var worker string // name of this instance to be used to know which instance run the job
	db.AutoMigrate(&CronJobLock{}) // We need the table to store the job execution
	locker, err := gormlock.NewGormLocker(db, worker)
	if err != nil {
		// handle the error
	}

	s := gocron.NewScheduler(time.UTC)
	s.WithDistributedLocker(locker)

	_, err = s.Every("1s").Name("unique_name").Do(func() {
		// task to do
		fmt.Println("call 1s")
	})
	if err != nil {
		// handle the error
	}

	s.StartBlocking()
}
```

## Prerequisites

- The table cron_job_locks needs to exist in the database. This can be achieved, as an example, using gorm automigrate functionality `db.Automigrate(&CronJobLock{})`
- In order to uniquely identify the job, the locker uses the unique combination of the job name + timestamp (by default with precision to seconds).

## FAQ

- Q: The locker uses the unique combination of the job name + timestamp with seconds precision, how can I change that?
    - A: It's possible to change the timestamp precision used to uniquely identify the job, here is an example to set an hour precision:
      ```go
      locker, err := gormlock.NewGormLocker(db, "local", gormlock.WithDefaultJobIdentifier(60 * time.Minute))
      ```
- Q: But what about if we want to write our own implementation:
    - A: It's possible to set how to create the job identifier:
      ```go
      locker, err := gormlock.NewGormLocker(db, "local",
          gormlock.WithJobIdentifier(
              func(ctx context.Context, key string) string {
                  return ...
              },
          ),
      )
      ```