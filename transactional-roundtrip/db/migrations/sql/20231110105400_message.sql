-- +goose Up
-- +goose StatementBegin
CREATE TABLE message(
  id varchar(27) NOT NULL PRIMARY KEY,

  producer_id text NOT NULL,
  consumer_id text NOT NULL,
  message_status SMALLINT NOT NULL,
  processing_status SMALLINT NOT NULL,

  inception_time TIMESTAMP WITH TIME ZONE NOT NULL,
  last_time TIMESTAMP WITH TIME ZONE NOT NULL,

  producer_replays SMALLINT NOT NULL DEFAULT 0,
  consumer_replays SMALLINT NOT NULL DEFAULT 0,

  operation_type SMALLINT NOT NULL,
  creditor_account text NOT NULL,
  debtor_account text NOT NULL,
  amount numeric(15,3) NOT NULL,
  currency char(3) NOT NULL,

  balance_before numeric(15,3) NULL,
  balance_after numeric(15,3) NULL,
  rejection_cause text NULL
);

CREATE INDEX idx_message_last_time ON message(last_time) 
INCLUDE(producer_id,consumer_id)
WHERE message_status < 3;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS message;
-- +goose StatementEnd
