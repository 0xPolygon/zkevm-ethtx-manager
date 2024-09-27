package sqlstorage

import (
	"database/sql"
	"embed"

	"github.com/0xPolygon/zkevm-ethtx-manager/log"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*
var dbMigrations embed.FS

// RunMigrations applies database migrations in the specified direction (up or down).
func RunMigrations(driverName string, db *sql.DB, direction migrate.MigrationDirection) error {
	migrations := migrate.EmbedFileSystemMigrationSource{
		FileSystem: dbMigrations,
		Root:       "migrations",
	}

	migrationsCount, err := migrate.Exec(db, driverName, migrations, direction)
	if err != nil {
		return err
	}

	log.Infof("Successfully ran %d migrations in direction: %v", migrationsCount, direction)
	return nil
}
