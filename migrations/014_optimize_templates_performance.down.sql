-- Rollback template performance optimizations
DROP INDEX IF EXISTS idx_forms_template_status;
DROP INDEX IF EXISTS idx_forms_updated_at;
