package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/kelseyhightower/envconfig"
	"github.com/lib/pq"

	"github.com/jmoiron/sqlx"
)

// Handler stores connection pool to database
type Handler struct {
	DB *sqlx.DB
}

// User is struct for (de)serializtion for user resourse from/to db and json
type User struct {
	ID        int64     `db:"id"         json:"id,omitempty"`
	Username  string    `db:"username"   json:"username"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
}

// Chat is struct for (de)serializtion for chat resourse from/to db and json
type Chat struct {
	ID            int64     `db:"id"              json:"id,omitempty"`
	Name          string    `db:"chat_name"       json:"name"`
	Users         []int64   `db:"-"               json:"users"`
	CreatedAt     time.Time `db:"created_at"      json:"created_at,omitempty"`
	LastMessageAt time.Time `db:"last_message_at" json:"last_message_at,omitempty"`
}

// Message is struct for (de)serializtion for message resourse from/to db and json
type Message struct {
	ID        int64     `db:"id"         json:"id,omitempty"`
	Chat      int64     `db:"chat_id"    json:"chat"`
	Author    int64     `db:"author_id"  json:"author"`
	Text      string    `db:"mes_text"   json:"text"`
	CreatedAt time.Time `db:"created_at" json:"created_at,omitempty"`
}

// ErrorMessage is struct for deserialization to json for error responses
type ErrorMessage struct {
	Error string `json:"error"`
}

// sendError sends error message in format {"error": "<message>"} to response writer w
func sendError(w http.ResponseWriter, statusCode int, errorMessage string) {
	w.WriteHeader(statusCode)
	errorStruct := ErrorMessage{
		Error: errorMessage,
	}
	errMsg, _ := json.Marshal(errorStruct)
	w.Write(errMsg)
}

// createUser takes json with single argument "username" and creates user with given username.
// Response codes:
// 	201:
// 		User created. No error.
// 	400:
//		User with given username already exists or argument "username" not given.
// 	500:
// 		Error in database. See error message for details.
// Body returns:
// 	User resource with status 201 or error message with others.
func (hnd *Handler) createUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	newUser := new(User)
	err := json.NewDecoder(r.Body).Decode(newUser)
	if err != nil || newUser.Username == "" {
		sendError(w, http.StatusBadRequest, err.Error())
	} else {

		// Probably it's unnecessary to send all fields in response,
		// but in some imaginary cases it can be useful (e.g. after registration
		// user can be redirected to his profile page, where is field with registration date)
		err = hnd.DB.Get(newUser, `
		INSERT INTO users_table (username) 
			VALUES ($1) 
			RETURNING id, username, created_at`, newUser.Username,
		)
		if err != nil {
			if pqErr, ok := err.(pq.Error); ok {
				if pqErr.Code == "23505" {
					sendError(w, http.StatusBadRequest, "User already exists")
					return
				}
			}
			sendError(w, http.StatusInternalServerError, err.Error())
		} else {
			response, _ := json.Marshal(newUser)
			w.WriteHeader(http.StatusCreated)
			w.Write(response)
		}
	}
}

// createChat takes json with single argument "name" and creates chat with given name.
// Response codes:
// 	201:
// 		Chat created. No error.
// 	400:
//		Chat with given name already exists or argument "name" not given.
// 	500:
// 		Error in database. See error message for details.
// Body returns:
// 	Chat resource with status 201 or error message with others.
func (hnd *Handler) createChat(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	newChat := new(Chat)
	err := json.NewDecoder(r.Body).Decode(newChat)
	if err != nil || newChat.Name == "" {
		sendError(w, http.StatusBadRequest, err.Error())
	} else {
		createChatTx, err := hnd.DB.Beginx()
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
		}

		err = createChatTx.Get(newChat, `
		INSERT INTO chats_table (chat_name) 
			VALUES ($1) 
			RETURNING id, chat_name, created_at`, newChat.Name,
		)
		if err != nil {
			defer createChatTx.Rollback()
			if pqErr, ok := err.(pq.Error); ok {
				if pqErr.Code == "23505" {
					sendError(w, http.StatusBadRequest, "Chat already exists")
					return
				}
			}
			sendError(w, http.StatusInternalServerError, err.Error())
			return
		}

		args := make([]interface{}, 0, len(newChat.Users)+1)
		args = append(args, newChat.ID)
		for _, user := range newChat.Users {
			args = append(args, user)
		}

		_, err = createChatTx.Exec(buildQueryAddUsersToChat(len(newChat.Users)), args...)
		if err != nil {
			sendError(w, http.StatusInternalServerError, err.Error())
			createChatTx.Rollback()
			return
		}

		createChatTx.Commit()
		response, _ := json.Marshal(newChat)
		w.WriteHeader(http.StatusCreated)
		w.Write(response)

	}
}

// sendMessage takes message resourse as argument and creates message to given chat from given user.
// Response codes:
// 	201:
// 		Message sent. No error.
// 	400:
//		Invalid message resourse given
// 	404:
//		Chat or author not found
// 	500:
// 		Error in database. See error message for details.
// Body returns:
// 	Full message resource with status 201 or error message with others.
func (hnd *Handler) sendMessage(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	newMes := new(Message)
	err := json.NewDecoder(r.Body).Decode(newMes)
	if err != nil || newMes.Author == 0 || newMes.Chat == 0 || newMes.Text == "" {
		sendError(w, http.StatusBadRequest, err.Error())
	} else {
		err = hnd.DB.Get(newMes, `
		INSERT INTO messages_table (chat_id, author_id, mes_text) 
			VALUES ($1, $2, $3)
			RETURNING id, chat_id, author_id, mes_text, created_at`, newMes.Chat, newMes.Author, newMes.Text,
		)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok {
				if pqErr.Code == "23503" {
					sendError(w, http.StatusNotFound, "User or chat not found")
					return
				}
			}
			sendError(w, http.StatusInternalServerError, err.Error())
		} else {
			response, _ := json.Marshal(newMes)
			w.WriteHeader(http.StatusCreated)
			w.Write(response)
		}
	}
}

type requestGetChats struct {
	User int64 `json:"user,required"`
	requestList
}

// getChats takes json with single argument "user" (its id) and returns list of chat resources.
// Also has optional arguments count and offset for limit number of returned chats.
// Response codes:
// 	200:
// 		No error.
// 	400:
//		Argument "user" not given.
// 	404:
// 		Given user not found.
// 	500:
// 		Error in database. See error message for details.
// Body returns:
// 	List of full chat resources with status 200 or error message with others.
func (hnd *Handler) getChats(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	request := new(requestGetChats)
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil || request.User == 0 {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	var found bool
	hnd.DB.Get(&found, `SELECT EXISTS(SELECT * FROM users_table WHERE id = $1)`, request.User)
	if !found {
		sendError(w, http.StatusNotFound, "User not found")
		return
	}

	query, args := buildQueryGetChats(request)

	type combinedResult struct {
		Chat        `db:"chat"`
		Participant int64 `db:"participant"`
	}

	var rows []combinedResult

	err = hnd.DB.Select(&rows, query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
	}

	chatIndex := make(map[int64]*Chat)
	var resChats []*Chat

	for _, row := range rows {
		ch, ok := chatIndex[row.Chat.ID]
		if !ok {
			ch = new(Chat)
			*ch = row.Chat
			ch.Users = make([]int64, 0)

			chatIndex[ch.ID] = ch
			resChats = append(resChats, ch)
		}
		ch.Users = append(ch.Users, row.Participant)
	}

	response, _ := json.Marshal(resChats)
	w.Write(response)
}

type requestGetMesages struct {
	Chat int64 `json:"chat,required"`
	requestList
}

// getMessages takes json with single argument "chat" (its id) and returns list of message resources.
// Also has optional arguments count and offset for limit number of returned chats.
// Response codes:
// 	200:
// 		No error.
// 	400:
//		Argument "chat" not given.
// 	404:
// 		Given chat not found.
// 	500:
// 		Error in database. See error message for details.
// Body returns:
// 	List of full message resources with status 200 or error message with others.
func (hnd *Handler) getMessages(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	request := new(requestGetMesages)
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	var found bool
	hnd.DB.Get(&found, `SELECT EXISTS(SELECT * FROM chats_table WHERE id = $1)`, request.Chat)
	if !found {
		sendError(w, http.StatusNotFound, "Chat not found")
		return
	}

	args := make([]interface{}, 0, 3)
	args = append(args, request.Chat)
	query := `
	SELECT id, chat_id, author_id, mes_text, created_at
		FROM messages_table
		WHERE chat_id = $1
		ORDER BY created_at DESC`

	if request.Count > 0 {
		var appConf appEnvSettings
		err := envconfig.Process("app", &appConf)
		if err != nil {
			appConf.MaxMessagesCount = 50
		}

		if request.Count > appConf.MaxMessagesCount {
			request.Count = appConf.MaxMessagesCount
		}

		if request.Offset < 0 {
			request.Offset = 0
		}

		query += "\nLIMIT $2\nOFFSET $3"
		args = append(args, request.Count)
		args = append(args, request.Offset)
	}

	var resMessages []*Message
	err = hnd.DB.Select(&resMessages, query, args...)
	if err != nil {
		sendError(w, http.StatusInternalServerError, err.Error())
	}

	response, _ := json.Marshal(resMessages)
	w.Write(response)
}

// initMux creates new router and links uri with functions
func initMux(hnd *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/users/add", restrictMethods([]string{"POST"}, hnd.createUser))
	mux.HandleFunc("/chats/add", restrictMethods([]string{"POST"}, hnd.createChat))
	mux.HandleFunc("/chats/get", restrictMethods([]string{"POST"}, hnd.getChats))
	mux.HandleFunc("/messages/add", restrictMethods([]string{"POST"}, hnd.sendMessage))
	mux.HandleFunc("/messages/get", restrictMethods([]string{"POST"}, hnd.getMessages))

	return mux
}
