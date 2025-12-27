-- Remove indexes
DROP INDEX IF EXISTS idx_forms_published_at;
DROP INDEX IF EXISTS idx_forms_custom_slug;
DROP INDEX IF EXISTS idx_forms_auto_slug;

-- Remove publish-related fields from forms table
ALTER TABLE forms
DROP COLUMN IF EXISTS published_at,
DROP COLUMN IF EXISTS accepting_responses,
DROP COLUMN IF EXISTS custom_slug,
DROP COLUMN IF EXISTS auto_slug;
