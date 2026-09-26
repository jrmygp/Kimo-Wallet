-- Migrations are forward-only and additive (docs/CLAUDE.md §5.4) — once a
-- migration has been applied anywhere, it is never edited; a correction is
-- always a new migration file.
CREATE TABLE transactions (
    id                 UUID PRIMARY KEY,
    idempotency_key    TEXT NOT NULL,
    -- No FK to another service's tables — a service owns its data
    -- (docs/CLAUDE.md §4 rule 1); these ids are owned by user-service /
    -- wallet-service, validated via gRPC at request time, not by a
    -- database constraint across services.
    sender_user_id     UUID NOT NULL,
    sender_wallet_id   UUID NOT NULL,
    receiver_user_id   UUID NOT NULL,
    receiver_wallet_id UUID NOT NULL,
    amount             BIGINT NOT NULL CHECK (amount > 0),
    currency           TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'PENDING',
    failure_reason     TEXT,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at       TIMESTAMPTZ,

    -- The idempotency guard (docs/CLAUDE.md §3.2 rule 3): a retried
    -- request for the same sender's same intent hits this constraint
    -- instead of creating a second transaction — see
    -- internal/storage/postgres/transaction_repository.go's Create,
    -- which maps a violation of this exact constraint to
    -- domain.ErrIdempotencyConflict.
    CONSTRAINT transactions_idempotency_key_sender_user_id_key UNIQUE (idempotency_key, sender_user_id)
);
