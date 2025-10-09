module examples

go 1.24.0

toolchain go1.24.1

replace github.com/go-co-op/gocron-gorm-lock/v2 => ../

require (
	github.com/go-co-op/gocron-gorm-lock/v2 v2.0.0-00010101000000-000000000000
	github.com/go-co-op/gocron/v2 v2.16.6
	gorm.io/driver/sqlite v1.6.0
	gorm.io/gorm v1.31.0
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	golang.org/x/text v0.27.0 // indirect
)
