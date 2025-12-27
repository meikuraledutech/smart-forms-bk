-- Add publish-related fields to forms table
ALTER TABLE forms
ADD COLUMN IF NOT EXISTS auto_slug VARCHAR(100) UNIQUE,
ADD COLUMN IF NOT EXISTS custom_slug VARCHAR(100) UNIQUE,
ADD COLUMN IF NOT EXISTS accepting_responses BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS published_at TIMESTAMP;

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_forms_auto_slug ON forms(auto_slug);
CREATE INDEX IF NOT EXISTS idx_forms_custom_slug ON forms(custom_slug);
CREATE INDEX IF NOT EXISTS idx_forms_published_at ON forms(published_at);
