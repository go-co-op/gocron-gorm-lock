# Gocron-Gorm-Lock

[![golangci-lint](https://github.com/go-co-op/gocron-gorm-lock/actions/workflows/go_test.yml/badge.svg)](https://github.com/go-co-op/gocron-gorm-lock/actions/workflows/go_test.yml)
![Go Report Card](https://goreportcard.com/badge/github.com/go-co-op/gocron-gorm-lock)
[![Go Doc](https://godoc.org/github.com/go-co-op/gocron-gorm-lock?status.svg)](https://pkg.go.dev/github.com/go-co-op/gocron-gorm-lock)

A gocron locker implementation using gorm

## ⬇️ Install

```bash
go get github.com/go-co-op/gocron-gorm-lock/v2
```

## 📋 Usage

Here is an example usage that would be deployed in multiple instances

```go
package main

import (
    "fmt"

    "github.com/go-co-op/gocron/v2"
    gormlock "github.com/go-co-op/gocron-gorm-lock/v2"
    "gorm.io/gorm"
    "time"
)

func main() {
    var db * gorm.DB // gorm db connection
    var worker string // name of this instance to be used to know which instance run the job 
    // db.AutoMigrate(&gormlock.CronJobLock{}) // We need the table to store the job execution
    locker, err := gormlock.NewGormLocker(db, worker)
    // handle the error
    
    s, err := gocron.NewScheduler(gocron.WithDistributedLocker(locker))
    // handle the error

    f := func() {
        // task to do
        fmt.Println("call 1s")
    }
    
    _, err = s.NewJob(gocron.DurationJob(1*time.Second), gocron.NewTask(f), gocron.WithName("unique_name"))
    if err != nil {
        // handle the error
    }

    s.Start()
}
```

To check a real use case example, check [examples](./examples).

## Prerequisites

- The table `cron_job_locks` needs to exist in the database. One possible option is to use [`db.Automigrate(&gormlock.CronJobLock{})`](https://gorm.io/docs/migration.html)
- In order to uniquely identify the job, the locker uses the unique combination of `job name + timestamp` (by default with precision to seconds). Check [JobIdentifier](#jobidentifier) for more info.

## 💡 Features

### JobIdentifier

JobIdentifier is how we identify when the job was run. 
Gorm Lock tries to lock the run of a job by uniquely identify a particular execution of a job with the combination of

* job name
* job identifier

The default implementation to uniquely identify a particular execution of job is using the following combination [`job name and timestamp`](./gorm_lock_options.go).

<details>
<summary>Example</summary>

Imagine that you have two instances running (`i1` and `i2`). 
And you configure a cron job (named `myJob`) to run at a certain period (e.g. every minute).

At `t1`, `i1` is faster in picking up the job, and then this happened: 

* `i1` creates a record in the database, (`jobName: test, jobIdentifier: t1`).
* Then, `i2` will try to lock and insert a record with the same values (`jobName: test, jobIdentifier: t1`). 
* But there is a combined unique constraint in the columns `jobName` and `jobIdentifier` making `i2` not able to run the job.
</details>

#### JobIdentifier Timestamp Precision

By default, the timestamp precision is in **seconds**, meaning that if a job named `myJob` is executed at `2025-01-01 10:11:12 15:16:17.000`, the resulting job identifier will be the combination of `myJob` and `2025-01-01 10:11:12`.

- It is possible to change the precision with [`WithDefaultJobIdentifier(newPrecision)`](./gorm_lock_options.go), e.g. `WithDefaultJobIdentifier(time.Hour)`
- It is also possible to completely override the way the job identifier is created with the [`WithJobIdentifier()`](./gorm_lock_options.go) option.

To see these two options in action, check the test [TestJobReturningExceptionWhenUnique](./gorm_lock_test.go)

### Removing Old Entries

Gorm Lock also removes old entries stored in `gormlock.CronJobLock`. You can configure the time interval, and the time to live with the following options:

- `WithTTL`
- `WithCleanInterval`
