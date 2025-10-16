-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE IF NOT EXISTS sessions(
    id VARCHAR(128) NOT NULL PRIMARY KEY ,
    user_id UUID REFERENCES users(id) NOT NULL ,
    is_active BOOL NOT NULL ,
    expires_at TIMESTAMP NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE IF EXISTS sessions;
-- +goose StatementEnd
