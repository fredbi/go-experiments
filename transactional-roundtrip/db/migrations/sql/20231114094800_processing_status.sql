-- +goose Up
-- +goose StatementBegin
ALTER TABLE message ADD COLUMN consumer_processing_status SMALLINT NOT NULL DEFAULT 0;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE message DROP COLUMN consumer_processing_status;
-- +goose StatementEnd
