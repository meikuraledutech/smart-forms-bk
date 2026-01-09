-- Optimize template queries performance
-- Add composite index for faster template listing
CREATE INDEX IF NOT EXISTS idx_forms_template_status ON forms(is_template, status, deleted_at)
WHERE is_template = true AND status = 'published' AND deleted_at IS NULL;

-- Add index on updated_at for ordering
CREATE INDEX IF NOT EXISTS idx_forms_updated_at ON forms(updated_at DESC);

COMMENT ON INDEX idx_forms_template_status IS 'Composite index for fast template listing queries';
COMMENT ON INDEX idx_forms_updated_at IS 'Index for ordering forms by update time';
