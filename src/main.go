package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

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
	Text      string    `db:"text"`
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
	var err error
	dbConf := new(dbSettings)
	err = envconfig.Process("db", dbConf)
	if err != nil {
		log.Fatalln(err)
	}
	dsnBuilder := buildDataSourceName(dbConf)

	hnd.DB, err = sqlx.Connect(dbConf.SQLDriver, dsnBuilder.String())
	if err != nil {
		log.Fatalf("Cannot connect to database: %v", err)
	}

	hnd.DB.SetMaxIdleConns(dbConf.MaxIdleConns)
	hnd.DB.SetMaxOpenConns(dbConf.MaxOpenConns)
	hnd.DB.SetConnMaxLifetime(dbConf.ConnMaxLifeTime)

	prepareQueries(hnd)
}

func buildDataSourceName(dbConf *dbSettings) *strings.Builder {
	dsnBuilder := strings.Builder{}
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

	return &dsnBuilder
}

func prepareQueries(hnd *Handler) {

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

	appConf := new(appSettings)
	err = envconfig.Process("app", appConf)
	if err != nil {
		log.Fatalln(err)
	}

	server := &http.Server{
		Addr:     appConf.ServerPort,
		ErrorLog: logger,
		Handler:  mux,
	}

	server.ListenAndServe()
}
