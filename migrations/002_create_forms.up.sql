-- Enable trigram extension (safe if already exists)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE forms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,

    title TEXT NOT NULL,
    description TEXT,

    status TEXT NOT NULL DEFAULT 'draft',

    created_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),

    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_forms_user_id ON forms(user_id);
CREATE INDEX idx_forms_updated_at ON forms(updated_at DESC);
CREATE INDEX idx_forms_title ON forms USING GIN (title gin_trgm_ops);
