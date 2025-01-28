package gormlock

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	gormMysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAutoMigration_MySQL(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		img  string
		args []string
	}{
		"8.0.36": {
			img:  "mysql:8.0.36",
			args: []string{"charset=utf8mb4", "parseTime=True", "loc=Local"},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			mysqlContainer, err := mysql.Run(ctx,
				tc.img,
				mysql.WithDatabase("test"),
				mysql.WithUsername("root"),
				mysql.WithPassword("password"),
			)
			require.NoError(t, err, "failed to start container")
			defer func() {
				if err := testcontainers.TerminateContainer(mysqlContainer); err != nil {
					t.Fatalf("failed to terminate container: %s", err)
				}
			}()

			connStr, errConn := mysqlContainer.ConnectionString(ctx, tc.args...)
			require.NoError(t, errConn, "failed to get mysql connection string")

			db, errOpeningConn := gorm.Open(gormMysql.Open(connStr), &gorm.Config{})
			require.NoError(t, errOpeningConn)

			err = db.AutoMigrate(&CronJobLock{})
			assert.NoError(t, err)
		})
	}
}
