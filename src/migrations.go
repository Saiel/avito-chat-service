package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/jmoiron/sqlx"
)

// MigrationsHandler stores configuration for migrations
type MigrationsHandler struct {
	WriteStdin         bool
	WriteStderr        bool
	MigrationTableName string
}

// InitMigrations creates table with trigger to track migrations
func (hnd *MigrationsHandler) InitMigrations(dbx *sqlx.DB) error {
	_, err := dbx.Exec(`
	CREATE TABLE IF NOT EXISTS ` + hnd.MigrationTableName + ` (
		migration_name varchar(80)
			PRIMARY KEY,
		created_at TIMESTAMP
			NOT NULL
	);

	CREATE OR REPLACE FUNCTION set_created_at()
		RETURNS TRIGGER
		LANGUAGE PLPGSQL
		AS $$
	BEGIN
		NEW.created_at = now();
		RETURN NEW;
	END;
	$$;

	DROP TRIGGER IF EXISTS migrated_at_trig ON ` + hnd.MigrationTableName + `;
	CREATE TRIGGER migrated_at_trig BEFORE INSERT
		ON migrations_chat_service
		FOR EACH ROW
		EXECUTE PROCEDURE set_created_at();
	`)
	if err != nil {
		hnd.errorf("Cannot initialize migrations: %v\n", err)
	}
	return err
}

// Migrate recursively executes all files in given directory if they did not executed before
func (hnd *MigrationsHandler) Migrate(dbx *sqlx.DB, pathToMigrations string) error {
	hnd.println("Starting migration...")
	err := hnd.migrate(dbx, pathToMigrations)
	if err == nil {
		hnd.println("Successfully migrated")
	} else {
		hnd.println("Migration stopped")
	}
	return err
}

func (hnd *MigrationsHandler) migrate(dbx *sqlx.DB, pathToMigrations string) error {
	fileInfos, err := ioutil.ReadDir(pathToMigrations)
	if err != nil {
		hnd.errorf("Cannot read path to migrations folder: %v\n", err)
		return err
	}

	for _, file := range fileInfos {
		hnd.printf("%v processing...\n", file.Name())

		if file.IsDir() {
			newMifrationFolder := path.Join(pathToMigrations, file.Name())
			err = hnd.migrate(dbx, newMifrationFolder) // reqursively migrate with subdirectories
			if err != nil {
				return err
			}
		} else {
			hnd.migrateFile(dbx, pathToMigrations, file.Name())
		}
	}
	return nil
}

func (hnd *MigrationsHandler) migrateFile(dbx *sqlx.DB, parentDir, fileName string) error {
	if !strings.HasSuffix(fileName, ".sql") {
		hnd.errorf("Not sql file\n")
		return fmt.Errorf("File %v is not a sql file", fileName)
	}
	migrationName := strings.TrimSuffix(fileName, ".sql")

	isMigrated, err := hnd.checkIfMigrated(dbx, migrationName)
	if err != nil {
		hnd.errorf("Error on checking migration: %v\n", err)
		return err
	}

	if isMigrated {
		hnd.println(" Already migrated")
	} else {
		data, err := ioutil.ReadFile(path.Join(parentDir, fileName))
		if err != nil {
			hnd.errorf("Cannot read file %v: %v\n", fileName, err)
			return err
		}

		err = hnd.performMigrate(dbx, migrationName, data)
		if err != nil {
			return err
		}
		hnd.println(" Success")
	}
	return nil
}

func (hnd *MigrationsHandler) performMigrate(dbx *sqlx.DB, migrationName string, queryBytes []byte) error {
	tx, err := dbx.Beginx()
	if err != nil {
		hnd.errorf("Migration: %v\n Cannot begin transaction: %v\n", migrationName, err)
		return err
	}

	buf := bytes.NewBuffer(queryBytes)
	_, err = tx.Exec(buf.String())
	if err != nil {
		hnd.errorf("Migration: %v\n Error on query: %v\n", migrationName, err)
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(
		`INSERT INTO `+hnd.MigrationTableName+` (migration_name) 
		VALUES ($1)`, migrationName,
	)
	if err != nil {
		hnd.errorf("Migration: %v\n Cannot record migration: %v\n", migrationName, err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		hnd.errorf("Migration: %v\n Cannot commit transaction: %v\n", migrationName, err)
		tx.Rollback()
		return err
	}
	return nil
}

func (hnd *MigrationsHandler) checkIfMigrated(dbx *sqlx.DB, fileName string) (bool, error) {
	// TODO: make less sql requests
	var res bool
	err := dbx.Get(&res, `
	SELECT EXISTS (
		SELECT * FROM `+hnd.MigrationTableName+` WHERE migration_name = $1
	)`, fileName)
	if err != nil {
		return false, err
	}
	return res, nil
}

func (hnd *MigrationsHandler) printf(format string, a ...interface{}) (int, error) {
	if hnd.WriteStdin {
		return fmt.Printf(format, a...)
	}
	return -1, nil
}

func (hnd *MigrationsHandler) println(a ...interface{}) (int, error) {
	if hnd.WriteStdin {
		return fmt.Println(a...)
	}
	return -1, nil
}

func (hnd *MigrationsHandler) errorf(format string, a ...interface{}) (int, error) {
	if hnd.WriteStderr {
		return fmt.Fprintf(os.Stderr, format, a...)
	}
	return -1, nil
}
