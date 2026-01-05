-- Add composite index on forms for user queries
-- Speeds up: WHERE user_id = X AND deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_forms_user_deleted
ON forms (user_id, deleted_at);

-- Add composite index on forms for ordering by update time
-- Speeds up: WHERE user_id = X AND deleted_at IS NULL ORDER BY updated_at DESC
CREATE INDEX IF NOT EXISTS idx_forms_user_updated
ON forms (user_id, deleted_at, updated_at DESC);

-- Add partial index for published form slug lookups
-- Speeds up: WHERE (auto_slug = X OR custom_slug = X) AND status = 'published' AND deleted_at IS NULL
-- Only indexes published, non-deleted forms (much smaller index)
CREATE INDEX IF NOT EXISTS idx_forms_published_slugs
ON forms (auto_slug, custom_slug)
WHERE status = 'published' AND deleted_at IS NULL;

-- Add composite index on flow_connections for form queries
-- Speeds up: WHERE form_id = X AND deleted_at IS NULL
CREATE INDEX IF NOT EXISTS idx_flow_connections_form_deleted
ON flow_connections (form_id, deleted_at);

-- Add composite index on form_responses for form queries
-- Speeds up: WHERE form_id = X ORDER BY submitted_at DESC
CREATE INDEX IF NOT EXISTS idx_form_responses_form_submitted
ON form_responses (form_id, submitted_at DESC);
