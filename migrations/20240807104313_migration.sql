-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD state INTEGER;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
