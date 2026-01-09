-- Remove template functionality
DROP INDEX IF EXISTS idx_forms_is_template;
ALTER TABLE forms DROP COLUMN IF EXISTS is_template;
