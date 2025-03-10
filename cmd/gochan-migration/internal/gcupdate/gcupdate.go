package gcupdate

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/gochan-org/gochan/cmd/gochan-migration/internal/common"
	"github.com/gochan-org/gochan/pkg/config"
	"github.com/gochan-org/gochan/pkg/gcsql"
	"github.com/gochan-org/gochan/pkg/gcutil"
)

const (
	// if the database version is less than this, it is assumed to be out of date, and the schema needs to be adjusted
	latestDatabaseVersion = 2
)

type GCDatabaseUpdater struct {
	options *common.MigrationOptions
	db      *gcsql.GCDB
}

func (dbu *GCDatabaseUpdater) Init(options *common.MigrationOptions) error {
	dbu.options = options
	criticalCfg := config.GetSystemCriticalConfig()
	var err error
	dbu.db, err = gcsql.Open(
		criticalCfg.DBhost, criticalCfg.DBtype, criticalCfg.DBname, criticalCfg.DBusername, criticalCfg.DBpassword,
		criticalCfg.DBprefix,
	)
	return err
}

func (dbu *GCDatabaseUpdater) IsMigrated() (bool, error) {
	var currentDatabaseVersion int
	err := dbu.db.QueryRowSQL(`SELECT version FROM DBPREFIXdatabase_version WHERE component = 'gochan'`, nil,
		[]any{&currentDatabaseVersion})
	if err != nil {
		return false, err
	}
	if currentDatabaseVersion == latestDatabaseVersion {
		return true, nil
	}
	if currentDatabaseVersion > latestDatabaseVersion {
		return false, fmt.Errorf("database layout is ahead of current version (%d), target version: %d",
			currentDatabaseVersion, latestDatabaseVersion)
	}
	return false, nil
}

func (dbu *GCDatabaseUpdater) MigrateDB() (bool, error) {
	migrated, err := dbu.IsMigrated()
	if migrated || err != nil {
		return migrated, err
	}

	var query string
	criticalConfig := config.GetSystemCriticalConfig()
	ctx := context.Background()
	tx, err := dbu.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: 0,
		ReadOnly:  false,
	})
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	switch criticalConfig.DBtype {
	case "mysql":
		query = `SELECT COUNT(*) FROM information_schema.TABLE_CONSTRAINTS
		WHERE CONSTRAINT_NAME = 'wordfilters_board_id_fk'
		AND TABLE_SCHEMA = DATABASE() AND TABLE_NAME = 'DBPREFIXwordfilters'`
		var numConstraints int

		if err = dbu.db.QueryRowTxSQL(tx, query, nil, []any{&numConstraints}); err != nil {
			return false, err
		}
		if numConstraints > 0 {
			query = `ALTER TABLE DBPREFIXwordfilters DROP FOREIGN KEY wordfilters_board_id_fk`
		} else {
			query = ""
		}
		query = `SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE()
		AND TABLE_NAME = 'DBPREFIXwordfilters'
		AND COLUMN_NAME = 'board_dirs'`
		var numColumns int
		if err = dbu.db.QueryRowTxSQL(tx, query, nil, []any{&numColumns}); err != nil {
			return false, err
		}
		if numColumns == 0 {
			query = `ALTER TABLE DBPREFIXwordfilters ADD COLUMN board_dirs varchar(255) DEFAULT '*'`
			if _, err = gcsql.ExecTxSQL(tx, query); err != nil {
				return false, err
			}
		}

		// Yay, collation! Everybody loves MySQL's default collation!
		query = `ALTER DATABASE ` + criticalConfig.DBname + ` CHARACTER SET = utf8mb4 COLLATE = utf8mb4_unicode_ci`
		if _, err = tx.Exec(query); err != nil {
			return false, err
		}

		query = `SELECT TABLE_NAME FROM information_schema.TABLES WHERE TABLE_SCHEMA = ?`
		rows, err := dbu.db.QuerySQL(query, criticalConfig.DBname)
		if err != nil {
			return false, err
		}
		defer rows.Close()
		for rows.Next() {
			var tableName string
			err = rows.Scan(&tableName)
			if err != nil {
				return false, err
			}
			query = `ALTER TABLE ` + tableName + ` CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci`
			if _, err = tx.Exec(query); err != nil {
				return false, err
			}
		}
		err = nil
	case "postgres":
		_, err = gcsql.ExecSQL(`ALTER TABLE DBPREFIXwordfilters DROP CONSTRAINT IF EXISTS board_id_fk`)
		if err != nil {
			return false, err
		}
		query = `ALTER TABLE DBPREFIXwordfilters ADD COLUMN IF NOT EXISTS board_dirs varchar(255) DEFAULT '*'`
		if _, err = dbu.db.ExecTxSQL(tx, query); err != nil {
			return false, err
		}
	case "sqlite3":
		_, err = gcsql.ExecSQL(`PRAGMA foreign_keys = ON`)
		if err != nil {
			return false, err
		}
		query = `SELECT COUNT(*) FROM PRAGMA_TABLE_INFO('DBPREFIXwordfilters') WHERE name = 'board_dirs'`
		var numColumns int
		if err = dbu.db.QueryRowSQL(query, nil, []any{&numColumns}); err != nil {
			return false, err
		}
		if numColumns == 0 {
			query = `ALTER TABLE DBPREFIXwordfilters ADD COLUMN board_dirs varchar(255) DEFAULT '*'`
			if _, err = dbu.db.ExecTxSQL(tx, query); err != nil {
				return false, err
			}
		}
	}

	query = `UPDATE DBPREFIXdatabase_version SET version = ? WHERE component = 'gochan'`
	_, err = dbu.db.ExecTxSQL(tx, query, latestDatabaseVersion)
	if err != nil {
		return false, err
	}
	return false, tx.Commit()
}

func (dbu *GCDatabaseUpdater) MigrateBoards() error {
	return gcutil.ErrNotImplemented
}

func (dbu *GCDatabaseUpdater) MigratePosts() error {
	return gcutil.ErrNotImplemented
}

func (dbu *GCDatabaseUpdater) MigrateStaff(password string) error {
	return gcutil.ErrNotImplemented
}

func (dbu *GCDatabaseUpdater) MigrateBans() error {
	return gcutil.ErrNotImplemented
}

func (dbu *GCDatabaseUpdater) MigrateAnnouncements() error {
	return gcutil.ErrNotImplemented
}

func (dbu *GCDatabaseUpdater) Close() error {
	if dbu.db != nil {
		return dbu.db.Close()
	}
	return nil
}
