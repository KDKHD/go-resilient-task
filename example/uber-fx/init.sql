DROP TABLE IF EXISTS tasks;

CREATE TABLE task
(
  id                     UUID PRIMARY KEY,
  type                   TEXT   NOT NULL,
  sub_type               TEXT NULL,
  status                 TEXT   NOT NULL,
  next_event_time        TIMESTAMPTZ(6) NOT NULL,
  state_time             TIMESTAMPTZ(3) NOT NULL,
  processing_client_id   TEXT NULL,
  processing_start_time  TIMESTAMPTZ(3) NULL,
  time_created           TIMESTAMPTZ(3) NOT NULL,
  time_updated           TIMESTAMPTZ(3) NOT NULL,
  processing_tries_count BIGINT NOT NULL,
  version                BIGINT NOT NULL,
  priority               INT    NOT NULL DEFAULT 5
);

CREATE INDEX task_idx1 ON task (status, next_event_time);

CREATE TABLE task_data
(
  task_id     UUID PRIMARY KEY NOT NULL,
  data_format INT              NOT NULL,
  data        BYTEA            NOT NULL
);

CREATE TABLE unique_task_key
(
  task_id  UUID PRIMARY KEY NOT NULL,
  key_hash INT              NOT NULL,
  key      TEXT             NOT NULL,
  unique (key_hash, key)
);

INSERT INTO task (id, type, status, next_event_time, state_time, time_created, time_updated, processing_tries_count, version)
VALUES ('00000000-0000-0000-0000-000000000001', 'payment_initiated', 'WAITING', '2025-01-01 00:00:00.000000', '2025-01-01 00:00:00.000', '2025-01-01 00:00:00.000', '2025-01-01 00:00:00.000', 0, 0);