package main

import (
	"bufio"
	"log"
	"net/http"
	"os"
	"time"
)

type Handler struct {
}

type User struct {
	ID        int
	Username  string
	CreatedAt time.Time
}

type Chat struct {
	ID        int
	Name      string
	Users     []*User
	CreatedAt time.Time
}

type Message struct {
	ID        int
	Chat      *Chat
	Author    *User
	Text      string
	CreatedAt time.Time
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

func main() {
	handler := &Handler{}
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
	server := &http.Server{
		Addr:     ":9000",
		ErrorLog: logger,
		Handler:  mux,
	}

	server.ListenAndServe()
}
