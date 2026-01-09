-- Add is_template field to forms
ALTER TABLE forms ADD COLUMN is_template BOOLEAN NOT NULL DEFAULT false;

-- Add index for template queries
CREATE INDEX idx_forms_is_template ON forms(is_template) WHERE is_template = true;

-- Add comment for clarity
COMMENT ON COLUMN forms.is_template IS 'Marks form as a template (super admin only)';
