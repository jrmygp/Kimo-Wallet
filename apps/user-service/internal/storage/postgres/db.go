package postgres

import (
	"fmt"

	gormpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open connects GORM to Postgres. TranslateError is on so driver-specific
// errors (e.g. a unique-constraint violation) come back wrapped as the
// underlying *pgconn.PgError we already check for in user_repository.go,
// instead of an opaque database/sql error.
//
// GORM's own logger is disabled: it writes plain-text SQL/error lines
// straight to stdout, which would bypass the service's structured JSON
// logging (docs/CLAUDE.md §5.4) and would log things like a duplicate phone
// number as an "ERROR" when it's actually an expected, handled outcome. The
// grpcserver/service layers already log what's worth logging.
func Open(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(gormpostgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	return db, nil
}
