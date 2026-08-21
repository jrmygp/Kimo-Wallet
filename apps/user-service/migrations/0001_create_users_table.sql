CREATE TABLE users (
    id            UUID PRIMARY KEY,
    phone_number  TEXT NOT NULL,
    full_name     TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT users_phone_number_key UNIQUE (phone_number)
);
