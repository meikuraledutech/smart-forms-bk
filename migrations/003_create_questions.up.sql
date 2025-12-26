CREATE TABLE IF NOT EXISTS questions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(20) NOT NULL CHECK (type IN ('question', 'option')),
    question_text TEXT NOT NULL,
    input_type VARCHAR(20),
    validation_rules JSONB DEFAULT '{}'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX idx_questions_type ON questions(type);
CREATE INDEX idx_questions_created_by ON questions(created_by);
CREATE INDEX idx_questions_created_at ON questions(created_at);
