CREATE TABLE IF NOT EXISTS users (
    id            UUID PRIMARY KEY,
    email         TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallets (
    user_id    UUID PRIMARY KEY REFERENCES users(id),
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transactions (
    id              UUID PRIMARY KEY,
    from_user_id    UUID NOT NULL REFERENCES users(id),
    to_user_id      UUID NOT NULL REFERENCES users(id),
    amount          BIGINT NOT NULL CHECK (amount > 0),
    notes           TEXT NOT NULL DEFAULT '',
    idempotency_key TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (from_user_id, idempotency_key),
    CHECK (from_user_id <> to_user_id)
);

CREATE INDEX IF NOT EXISTS idx_tx_from ON transactions(from_user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_tx_to   ON transactions(to_user_id, created_at DESC);