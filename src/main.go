package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Saiel/avito-chat-serivce/src/lib/migrations"

	"github.com/kelseyhightower/envconfig"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Handler struct {
	DB      *sqlx.DB
	queries map[string]*sqlx.Stmt
}

type User struct {
	ID        int       `db:"id"`
	Username  string    `db:"username"`
	CreatedAt time.Time `db:"created_at"`
}

type Chat struct {
	ID        int       `db:"id"`
	Name      string    `db:"chat_name"`
	Users     []*User   `db:"users"`
	CreatedAt time.Time `db:"created_at"`
}

type Message struct {
	ID        int       `db:"id"`
	Chat      *Chat     `db:"chat"`
	Author    *User     `db:"author"`
	Text      string    `db:"mes_text"`
	CreatedAt time.Time `db:"created_at"`
}

func (hnd *Handler) createUser(w http.ResponseWriter, r *http.Request) {

}

func (hnd *Handler) createChat(w http.ResponseWriter, r *http.Request) {

}

func (hnd *Handler) sendMessage(w http.ResponseWriter, r *http.Request) {

}

func (hnd *Handler) getChats(w http.ResponseWriter, r *http.Request) {

}

func (hnd *Handler) getMessages(w http.ResponseWriter, r *http.Request) {

}

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

	// prepareQueries(hnd)
}

func initConnection(hnd *Handler, dbConf *dbEnvSettings) error {
	dsn := buildDataSourceName(dbConf)

	db, err := sqlx.Connect(dbConf.SQLDriver, dsn)
	if err != nil {
		return err
	}

	hnd.DB = db
	hnd.DB.SetMaxIdleConns(dbConf.MaxIdleConns)
	hnd.DB.SetMaxOpenConns(dbConf.MaxOpenConns)
	hnd.DB.SetConnMaxLifetime(dbConf.ConnMaxLifeTime)
	return nil
}

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

func prepareQueries(hnd *Handler) {
	hnd.queries = make(map[string]*sqlx.Stmt)
	hnd.queries["create-user"] = prepareStmt(hnd, `INSERT INTO users_table (username) VALUES ($1)`)
	hnd.queries["create-chat"] = prepareStmt(hnd, `INSERT INTO chats_table (chat_name) VALUES ($1)`)
	hnd.queries["add-user-to-chat"] = prepareStmt(hnd, `INSERT INTO chats_users_table (chat_id, user_id) VALUES ($1, $2)`)
	hnd.queries["create-message"] = prepareStmt(hnd, `INSERT INTO messages_table (chat, author, text) VALUES ($1, $2, $3)`)
	hnd.queries["get-chats-of-user"] = prepareStmt(hnd, `SELECT id, name, created_at 
	FROM chats_table AS ch, users_table AS us, chats_users_table AS ch_us 
	WHERE ch.id = ch_us.chat AND us.id = ch_us.user`)
	// hnd.queries["get-messages-from-chat"] = prepareStmt(hnd, `SELECT `)
}

func prepareStmt(hnd *Handler, query string) *sqlx.Stmt {
	st, err := hnd.DB.Preparex(query)
	if err != nil {
		log.Fatalf("Cannot prepare statement: '%v'\nError: %v", query, err)
	}
	return st
}

func main() {
	handler := &Handler{}
	initDB(handler)
	errorFile, err := os.Create("error.log")
	if err != nil {
		panic(err)
	}
	logWriter := bufio.NewWriter(errorFile)
	logger := log.New(logWriter, "", 0)
	mux := http.NewServeMux()
	mux.HandleFunc("/users/add", handler.createUser)
	mux.HandleFunc("/chats/add", handler.createChat)
	mux.HandleFunc("/chats/get", handler.getChats)
	mux.HandleFunc("/messages/add", handler.sendMessage)
	mux.HandleFunc("/messages/get", handler.getMessages)

	appConf := new(appEnvSettings)
	err = envconfig.Process("app", appConf)
	if err != nil {
		log.Fatalln(err)
	}

	server := &http.Server{
		Addr:     ":" + appConf.ServerPort,
		ErrorLog: logger,
		Handler:  mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Cannot serve server: %v", err)
	}
}
