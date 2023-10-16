CREATE TABLE idempotency_headers (
    idempotency_key TEXT NOT NULL,
    header_name TEXT,
    header_value BYTEA,
    PRIMARY KEY (idempotency_key)
);
