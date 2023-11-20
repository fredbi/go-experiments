-- +goose Up
-- +goose StatementBegin
CREATE TABLE process_audit(
  audit_id serial NOT NULL PRIMARY KEY,
  id varchar(27) NOT NULL,
  processing_status SMALLINT NOT NULL,
  action text NULL,
  ts timestamp with time zone NOT NULL DEFAULT current_timestamp
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS process_audit;
-- +goose StatementEnd
