-- TEMPLATE NOTE: rename this file/table/columns to match your real entity.
-- Migrations are forward-only and additive (docs/CLAUDE.md §5.4) — once a
-- migration has been applied anywhere, it is never edited; a correction is
-- always a new migration file.
CREATE TABLE widgets (
    id         UUID PRIMARY KEY,
    name       TEXT NOT NULL,
    status     TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
