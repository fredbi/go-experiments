-- +goose Up
-- +goose StatementBegin
CREATE TABLE message(
  id serial NOT NULL PRIMARY KEY,
  producer_id SMALLINT NOT NULL,
  consumer_id SMALLINT NOT NULL,
  message_status SMALLINT NOT NULL,
  processing_status SMALLINT NOT NULL,

  inception_time TIMESTAMP WITH TIMEZONE NOT NULL,
  last_time TIMESTAMP WITH TIMEZONE NOT NULL,

  producer_replays SMALLINT NOT NULL DEFAULT 0,
  consumer_replays SMALLINT NOT NULL DEFAULT 0,

  operation_name TEXT NOT NULL DEFAULT '',
  result TEXT NOT NULL DEFAULT ''
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS message;
-- +goose StatementEnd
