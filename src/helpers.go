package main

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// requestList is struct with fields for limiting number of items in returned list
// Uses as parent struct
type requestList struct {
	Count  int `json:"count,omitempty"`
	Offset int `json:"offset,omitempty"`
}

func buildQueryAddUsersToChat(userCount int) string {
	builder := &strings.Builder{}
	builder.Grow(512)
	builder.WriteString(`INSERT INTO chats_users_table (chat_id, user_id) VALUES `)
	for i := 0; i < userCount; i++ {
		buildChatUserValueStr(i, builder)
		if i != userCount-1 {
			builder.WriteString(", ")
		}
	}

	return builder.String()
}

func buildChatUserValueStr(i int, builder *strings.Builder) {
	builder.WriteString("($1, $")
	builder.WriteString(strconv.Itoa(i + 2))
	builder.WriteString(")")
}

// BuildQueryGetChats builds query
// `SELECT
// 		chs.id              as "chat.id",
// 		chs.chat_name       as "chat.chat_name",
// 		chs.last_message_at as "chat.last_message_at",
// 		chs.created_at      as "chat.created_at",
// 		cu.user_id          as "participant"
// FROM chats_users_table as cu
// 		INNER JOIN (
// 			SELECT ch.id, ch.chat_name, ch.last_message_at, ch.created_at
// 				FROM chats_table ch
// 					INNER JOIN chats_users_table ch_us
// 						ON ch_us.chat_id = ch.id
// 				WHERE ch_us.user_id = $1
// 				ORDER BY last_message_at DESC
// 				[LIMIT $2
// 				OFFSET $3]
// 			) as chs on cu.chat_id = r.id`
//
// and returns arguments for it
func buildQueryGetChats(request *requestGetChats) (query string, args []interface{}) {
	args = make([]interface{}, 0, 3)
	args = append(args, request.User)

	builder := &strings.Builder{}
	builder.Grow(1024)
	builder.WriteString(`
	SELECT 
			chs.id              as "chat.id",
			chs.chat_name       as "chat.chat_name",
			chs.last_message_at as "chat.last_message_at",
			chs.created_at      as "chat.created_at",
			cu.user_id          as "participant"
		FROM chats_users_table as cu
			INNER JOIN (
				SELECT ch.id, ch.chat_name, ch.last_message_at, ch.created_at
					FROM chats_table ch
						INNER JOIN chats_users_table ch_us
							ON ch_us.chat_id = ch.id
					WHERE ch_us.user_id = $1
					ORDER BY last_message_at DESC`)

	if request.Count > 0 {
		var appConf appEnvSettings
		err := envconfig.Process("app", &appConf)
		if err != nil {
			appConf.MaxChatsCount = 20
		}

		if request.Count > appConf.MaxChatsCount {
			request.Count = appConf.MaxChatsCount
		}

		if request.Offset < 0 {
			request.Offset = 0
		}

		builder.WriteString("\nLIMIT $2\nOFFSET $3")
		args = append(args, request.Count)
		args = append(args, request.Offset)
	}

	builder.WriteString(`) as chs on cu.chat_id = chs.id`)

	return builder.String(), args
}

// restrictMethods is a middleware, which prohibit all http methods for handler that not in given slice
func restrictMethods(methods []string, f http.HandlerFunc) http.HandlerFunc {
	sort.Strings(methods)
	return func(w http.ResponseWriter, r *http.Request) {
		i := sort.SearchStrings(methods, r.Method)
		if i < len(methods) && methods[i] == r.Method {
			f(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
