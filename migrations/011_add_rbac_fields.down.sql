-- Drop subscription_plans table
DROP TABLE IF EXISTS subscription_plans;

-- Remove template fields from forms table
DROP INDEX IF EXISTS idx_forms_is_template;
ALTER TABLE forms DROP COLUMN IF EXISTS template_description;
ALTER TABLE forms DROP COLUMN IF EXISTS template_category;
ALTER TABLE forms DROP COLUMN IF EXISTS is_template;

-- Remove role from users table
DROP INDEX IF EXISTS idx_users_role;
ALTER TABLE users DROP COLUMN IF EXISTS role;
