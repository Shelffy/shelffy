-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE users
    DROP COLUMN IF EXISTS is_admin;
-- +goose StatementEnd
