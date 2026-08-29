-- profile_picture: nullable from the start, no backfill needed.
ALTER TABLE users ADD COLUMN profile_picture TEXT;

-- kimo_id: must end up NOT NULL and UNIQUE, but this table may already
-- have rows (existing users have no kimo_id yet). Add it nullable first,
-- backfill every existing row with a freshly generated value, then apply
-- the NOT NULL / UNIQUE constraints once no NULLs remain — doing it in the
-- opposite order would fail outright on any pre-existing row.
ALTER TABLE users ADD COLUMN kimo_id VARCHAR(12);

DO $$
DECLARE
    row_id    UUID;
    candidate CHAR(12);
    alphabet  CONSTANT TEXT := 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789';
BEGIN
    FOR row_id IN SELECT id FROM users WHERE kimo_id IS NULL LOOP
        LOOP
            SELECT string_agg(substr(alphabet, (random() * length(alphabet))::int + 1, 1), '')
            INTO candidate
            FROM generate_series(1, 12);

            EXIT WHEN NOT EXISTS (SELECT 1 FROM users WHERE kimo_id = candidate);
        END LOOP;

        UPDATE users SET kimo_id = candidate WHERE id = row_id;
    END LOOP;
END $$;

ALTER TABLE users ALTER COLUMN kimo_id SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_kimo_id_key UNIQUE (kimo_id);
