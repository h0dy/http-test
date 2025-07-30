-- +goose Up
CREATE TABLE refresh_tokens(
    token TEXT PRIMARY KEY NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expire_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP
);

-- +goose Down 
DROP TABLE refresh_tokens;
