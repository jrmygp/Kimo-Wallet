CREATE TABLE wallets (
    id         UUID PRIMARY KEY,
    -- The user-service `id` this wallet belongs to. No foreign key: this
    -- service owns its own database and never references another
    -- service's tables directly (docs/CLAUDE.md §4 rule 1) — referential
    -- integrity here comes from the flow that creates the row (consuming
    -- user.created), not a database constraint across services.
    user_id    UUID NOT NULL,
    balance    BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0),
    currency   TEXT NOT NULL DEFAULT 'IDR',
    -- No CHECK-constrained enum of allowed values yet — no status
    -- lifecycle is documented anywhere to encode as one (see
    -- docs/guides/Kimo-Wallet-Architecture.md / PRD, which only ever
    -- mention "wallet status" as a concept, no concrete values).
    status     TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- One wallet per user. Also the natural idempotency guard for the
    -- user.created consumer: a redelivered event just hits this
    -- constraint and is treated as an already-handled no-op, not a
    -- separate processed-events table (see internal/service's WalletService).
    CONSTRAINT wallets_user_id_key UNIQUE (user_id)
);
