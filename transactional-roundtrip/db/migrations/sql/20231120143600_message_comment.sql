-- +goose Up
-- +goose StatementBegin
ALTER TABLE message ADD COLUMN comment text;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE message DROP COLUMN comment;
-- +goose StatementEnd
