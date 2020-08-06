CREATE TABLE IF NOT EXISTS users_table (
    id         BIGINT
        GENERATED ALWAYS AS IDENTITY
        PRIMARY KEY,                        -- index
    username   VARCHAR(20)
        NOT NULL
        CONSTRAINT username_unique UNIQUE,  -- index
    created_at TIMESTAMP
        NOT NULL
);


CREATE TABLE IF NOT EXISTS chats_table (
    id         BIGINT
        GENERATED ALWAYS AS IDENTITY
        PRIMARY KEY,                        -- index
    chat_name  VARCHAR(40)
        NOT NULL
        CONSTRAINT chat_name_unique UNIQUE, -- index
    last_message_at TIMESTAMP
        DEFAULT now(),
    created_at TIMESTAMP
        NOT NULL
);


CREATE TABLE IF NOT EXISTS chats_users_table (
    user_id BIGINT
        NOT NULL
        REFERENCES users_table(id)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    chat_id BIGINT
        NOT NULL
        REFERENCES chats_table(id)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    PRIMARY KEY (user_id, chat_id)
);


CREATE TABLE IF NOT EXISTS messages_table (
    id         BIGINT
        GENERATED ALWAYS AS IDENTITY
        PRIMARY KEY,                        -- index
    chat_id    BIGINT
        NOT NULL
        REFERENCES chats_table(id)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    author_id  BIGINT
        NOT NULL
        REFERENCES users_table(id)
            ON DELETE CASCADE
            ON UPDATE CASCADE,
    mes_text   TEXT
        NOT NULL,
    created_at TIMESTAMP
        NOT NULL
);
