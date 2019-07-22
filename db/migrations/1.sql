-- +migrate Up
CREATE TABLE IF NOT EXISTS devices (
    id                 serial,
    namespace          text NOT NULL,
    device_id          text NOT NULL,
    device_uri         text NOT NULL,
    session_timeout    int NOT NULL DEFAULT 120,
    ping_interval      int NOT NULL DEFAULT 104,
    pong_timeout       int NOT NULL DEFAULT 16,
    events_topic       text NOT NULL DEFAULT 'deviceevent',
    created_at         timestamp NOT NULL DEFAULT now(),
    updated_at         timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS events (
    id             serial,
    namespace      text NOT NULL,
    source_type    text NOT NULL,
    source_id      text NOT NULL,
    topic          text NOT NULL,
    timestamp      timestamp NOT NULL DEFAULT now(),
    details        text NOT NULL,
    created_at     timestamp NOT NULL DEFAULT now(),
    updated_at     timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS sessions (
    id                 serial,
    namespace          text NOT NULL,
    device_id          text NOT NULL,
    device_uri         text NOT NULL,
    session_timeout    int NOT NULL DEFAULT 120,
    last_message_at    timestamp NOT NULL DEFAULT now(),
    created_at         timestamp NOT NULL DEFAULT now(),
    updated_at         timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY (id)
);

-- +migrate Down
DROP TABLE sessions;
DROP TABLE events;
DROP TABLE devices;
