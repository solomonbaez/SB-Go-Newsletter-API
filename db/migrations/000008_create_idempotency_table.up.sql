CREATE TABLE idempotency (
    id uuid NOT NULL REFERENCES users(id),
    idempotency_key TEXT NOT NULL,
    response_status_code SMALLINT NOT NULL,
    response_body BYTEA NOT NULL,
    created timestamptz NOT NULL,
    PRIMARY KEY(id, idempotency_key)
);