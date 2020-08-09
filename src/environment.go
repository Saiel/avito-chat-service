package main

import "time"

type dbEnvSettings struct {
	SQLDriver       string        `required:"true"`
	Host            string        `required:"true"`
	Port            string        `required:"true"`
	Name            string        `required:"true"`
	User            string        `required:"true"`
	Pass            string        `required:"true"`
	MaxIdleConns    int           `default:"4"`
	MaxOpenConns    int           `default:"16"`
	ConnMaxLifeTime time.Duration `default:"100m"`
	MigrationsDir   string        `default:"./migrations"`
}

type appEnvSettings struct {
	ServerPort       string `required:"true"`
	MaxChatsCount    int    `default:"20"`
	MaxMessagesCount int    `default:"50"`
}
