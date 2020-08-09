package main

import (
	"log"
	"strings"

	"github.com/Saiel/avito-chat-serivce/src/lib/migrations"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
)

// initDB connects to database with given config as environment variables with prefix DB_
// and handles new migrations
func initDB(hnd *Handler) {
	dbConf := new(dbEnvSettings)
	err := envconfig.Process("db", dbConf)
	if err != nil {
		log.Fatalln(err)
	}

	err = initConnection(hnd, dbConf)
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	migrationsHnd := &migrations.Handler{
		WriteStdin:         true,
		WriteStderr:        true,
		MigrationTableName: "migrations_chat_service",
	}
	err = migrationsHnd.InitMigrations(hnd.DB)
	if err != nil {
		log.Fatalln()
	}
	err = migrationsHnd.Migrate(hnd.DB, dbConf.MigrationsDir)
	if err != nil {
		log.Fatalln()
	}
}

// initConnections directly connects to db and returns error, if something went wrong
func initConnection(hnd *Handler, dbConf *dbEnvSettings) error {
	dsn := buildDataSourceName(dbConf)

	db, err := sqlx.Connect(dbConf.SQLDriver, dsn)
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(dbConf.MaxIdleConns)
	db.SetMaxOpenConns(dbConf.MaxOpenConns)
	db.SetConnMaxLifetime(dbConf.ConnMaxLifeTime)
	hnd.DB = db
	return nil
}

// buildDataSourceName build connection string
func buildDataSourceName(dbConf *dbEnvSettings) string {
	dsnBuilder := &strings.Builder{}
	dsnBuilder.Grow(256)

	dsnBuilder.WriteString("host=")
	dsnBuilder.WriteString(dbConf.Host)

	dsnBuilder.WriteString(" port=")
	dsnBuilder.WriteString(dbConf.Port)

	dsnBuilder.WriteString(" dbname=")
	dsnBuilder.WriteString(dbConf.Name)

	dsnBuilder.WriteString(" user=")
	dsnBuilder.WriteString(dbConf.User)

	dsnBuilder.WriteString(" password=")
	dsnBuilder.WriteString(dbConf.Pass)

	dsnBuilder.WriteString(" sslmode=disable")

	return dsnBuilder.String()
}
