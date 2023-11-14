-- +goose Up
-- +goose StatementBegin
ALTER TABLE message ADD COLUMN consumer_message_status SMALLINT NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE message DROP COLUMN consumer_message_status;
-- +goose StatementEnd
