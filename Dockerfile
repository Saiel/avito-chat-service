FROM golang:1.14-alpine

WORKDIR /app

COPY src/go.sum src/go.mod ./
RUN go mod download

COPY src/*.go src/*.sh ./
RUN go build -o main .

COPY src/migrations/ ./migrations

EXPOSE 9000

CMD [ "./start.sh", "./main" ]
