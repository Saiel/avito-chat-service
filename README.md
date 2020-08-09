# Simple chat service

Service implements basic logic for chats with several users.

Developed as trainee task from avito

## Implementation

### Starting server

```shell
$ sudo docker-compose up
```

### Technologies used 

* Docker with docker-compose
* PostgreSQL
* Golang

### Dependencies:

* github.com/jmoiron/sqlx
* github.com/lib/pq
* github.com/kelseyhightower/envconfig

### Implementations details

* Implemented basic **migration system** (in [`migrations.go`](src/migrations.go))
* * `MigrationHandler.InitiMigrations(*sqlx.DB)` generates migration table
* * `MigrationsHandler.Migrate(*sqlx.DB, pathToMigrations string)` reads all .sql files in [`pathToMigrations`](src/migrations/) and aply new ones.
* Tables and triggers are generated on first server startup via migrations
* Database is generating on image build with database and user, provided in environment variables
* Database connection inits in [`database.go`](src/database.go) file
* API logic implemented in [`http.go`](src/http.go)
* [`helpers.go`](src/helpers.go) stores additional functions
* For environment variables handling used [envconfig](https://github.com/kelseyhightower/envconfig) just for ease of use.
* Sorting of user's chats implements via field `last_message_at` with trigger on updating messages table
* **Pagination** in list responses via optional `count` and `offset` fields

### Further possible improvemnts

* **Tests**
* **Benchmarks**
* Long polling for handling new messages
* User authorization
* Json serialization via codegen (e.g. [easyjson](https://github.com/mailru/easyjson))
* Adding/deleting users to/from chat
* Parrallel migration of files with same prefix
* Verbose logging of incoming requests
* Other optimizations I forgot to implement in-place

## Task
Chat server that provides HTTP API for working with chats and users' messages.

### Requirements
* Any language
* Any techology for storing data
* Data must be stored between server restarts
* Server must be exposed on port 9000
* GUI is not nessesary, but not restricted.
* Give instruction for start server. Ideally: `docker-compose up`

### Main entities

#### User

Application user. Has this fields:

* id - unique identificator (can be as number as string)
* username - unique name of user
* created_at - time of user creation

#### Chat

Separate chat. Has this fields:

* id - unique chat identificator
* name - unique chat name
* users - list of users in chat. Many-to-many relation
* created_at - time of creation

#### Message

Message in chat. Has this fields:

* id - unique identificator of message
* chat - reference on chat's id, which message was sent in
* author - reference on author's id. Relation many-to one
* text - text of sended message
* created_at - time of creation

### Base API nethods

Methods proces HTTP POST requests with all required parameters in JSON body.

### Create new user

Request:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"username": "user_1"}' \
  http://localhost:9000/users/add
```

Response: `id` of created user or HTTP error code + error description.

### Create new chat with users

Request:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"name": "chat_1", "users": ["<USER_ID_1>", "<USER_ID_2>"]}' \
  http://localhost:9000/chats/add
```

Response: `id` of created chat or HTTP error code or HTTP error code + error description.

Amount of users in chat is not limited.

### Send message to chat on behalf of the user

Request:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"chat": "<CHAT_ID>", "author": "<USER_ID>", "text": "hi"}' \
  http://localhost:9000/messages/add
```

Response: `id` of created message or HTTP error code + error description.

### Get list of chats of concrete user

Request:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"user": "<USER_ID>"}' \
  http://localhost:9000/chats/get
```

Response: list of all user's chats with all fields sorted by last message sended in chat (from last to first) or HTTP error code + error description.

### Get list of messages in concrete chat

Request:

```bash
curl --header "Content-Type: application/json" \
  --request POST \
  --data '{"chat": "<CHAT_ID>"}' \
  http://localhost:9000/messages/get
```

Response: list of all messages in chat with all fields sorted by sending date (from last to first) or HTTP error code + error description.
