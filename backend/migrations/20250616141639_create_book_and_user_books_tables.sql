-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS books(
    id uuid primary key,
    title TEXT NOT NULL DEFAULT 'new book',
    uploaded_by uuid references users(id) NOT NULL,
    uploaded_at timestamp default now(),
    hash bytea NOT NULL, -- SHA-256
    path TEXT NOT NULL
);

CREATE OR REPLACE FUNCTION handle_book_unique_title()
    RETURNS TRIGGER AS $$
DECLARE
    base_title TEXT := NEW.title;
    new_title TEXT := base_title;
    counter INT := 1;
    title_exists BOOLEAN;
BEGIN
    SELECT EXISTS (
        SELECT 1
        FROM books
        WHERE title = new_title
          AND uploaded_by = NEW.uploaded_by
          AND (TG_OP = 'INSERT' OR id != NEW.id)
    ) INTO title_exists;

    WHILE title_exists LOOP
            new_title := base_title || ' (' || counter || ')';
            SELECT EXISTS (
                SELECT 1
                FROM books
                WHERE title = new_title
                  AND uploaded_by = NEW.uploaded_by
                  AND (TG_OP = 'INSERT' OR id != NEW.id)
            ) INTO title_exists;
            counter := counter + 1;
        END LOOP;

    NEW.title := new_title;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
DROP TRIGGER IF EXISTS unique_name_trigger_insert ON books;
CREATE TRIGGER unique_name_trigger_insert
    BEFORE INSERT ON books
    FOR EACH ROW
EXECUTE FUNCTION handle_book_unique_title();

DROP TRIGGER IF EXISTS unique_name_trigger_update ON books;
CREATE TRIGGER unique_name_trigger_update
    BEFORE UPDATE OF title, uploaded_by ON books
    FOR EACH ROW
EXECUTE FUNCTION handle_book_unique_title();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TRIGGER IF EXISTS unique_name_trigger_update ON books;
DROP TRIGGER IF EXISTS unique_name_trigger_insert ON books;
DROP TABLE IF EXISTS books;
-- +goose StatementEnd
