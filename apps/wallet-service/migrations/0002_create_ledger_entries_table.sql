CREATE TABLE ledger_entries (
    id             UUID PRIMARY KEY,
    transaction_id UUID NOT NULL,
    wallet_id      UUID NOT NULL,
    direction      TEXT NOT NULL CHECK (direction IN ('DEBIT','CREDIT')),
    amount         BIGINT NOT NULL CHECK (amount > 0),
    currency       TEXT NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT ledger_entries_transaction_id_direction_key UNIQUE (transaction_id, direction)
);