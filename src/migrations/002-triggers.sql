CREATE OR REPLACE FUNCTION update_last_message_timestamp()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
    AS $$
BEGIN
    UPDATE chats_table
        SET last_message_at = NEW.created_at
        WHERE NEW.chat_id = chats_table.id;
    RETURN NEW;
END;
$$;

CREATE TRIGGER update_last_message_timestamp_trig AFTER INSERT
    ON messages_table
    FOR EACH ROW
    EXECUTE PROCEDURE update_last_message_timestamp();


CREATE OR REPLACE FUNCTION set_created_at()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
    AS $$
BEGIN
    NEW.created_at = now();
    RETURN NEW;
END;
$$;

CREATE TRIGGER messages_created_at_trig BEFORE INSERT
    ON messages_table
    FOR EACH ROW
    EXECUTE PROCEDURE set_created_at();

CREATE TRIGGER users_created_at_trig BEFORE INSERT
    ON users_table
    FOR EACH ROW
    EXECUTE PROCEDURE set_created_at();

CREATE TRIGGER chats_created_at_trig BEFORE INSERT
    ON chats_table
    FOR EACH ROW
    EXECUTE PROCEDURE set_created_at();
