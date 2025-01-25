module examples

go 1.22

replace github.com/go-co-op/gocron-gorm-lock/v2 => ../

require (
	github.com/go-co-op/gocron-gorm-lock/v2 v2.0.0-00010101000000-000000000000
	github.com/go-co-op/gocron/v2 v2.15.0
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	golang.org/x/exp v0.0.0-20240613232115-7f521ea00fb8 // indirect
	golang.org/x/text v0.21.0 // indirect
)
