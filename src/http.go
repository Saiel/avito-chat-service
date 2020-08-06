package main

import (
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
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

func initMux(hnd *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/add", hnd.createUser)
	mux.HandleFunc("/chats/add", hnd.createChat)
	mux.HandleFunc("/chats/get", hnd.getChats)
	mux.HandleFunc("/messages/add", hnd.sendMessage)
	mux.HandleFunc("/messages/get", hnd.getMessages)

	return mux
}
