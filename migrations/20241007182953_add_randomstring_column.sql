-- +goose Up
-- +goose StatementBegin
ALTER TABLE times ADD randomString TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
