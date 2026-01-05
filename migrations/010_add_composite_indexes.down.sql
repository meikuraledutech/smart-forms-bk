-- Drop composite indexes
DROP INDEX IF EXISTS idx_forms_user_deleted;
DROP INDEX IF EXISTS idx_forms_user_updated;
DROP INDEX IF EXISTS idx_forms_published_slugs;
DROP INDEX IF EXISTS idx_flow_connections_form_deleted;
DROP INDEX IF EXISTS idx_form_responses_form_submitted;
