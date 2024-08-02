-- +goose Up
-- +goose StatementBegin
CREATE TABLE users(
    id INTEGER NOT NULL PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    onetimecode TEXT,
    authcode TEXT
);

CREATE TABLE times(
    id INTEGER NOT NULL PRIMARY KEY,
    userid INTEGER REFERENCES users (id),
    starttime INTEGER NOT NULL,
    endtime INTEGER NULL,
    computedtime INTEGER
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
