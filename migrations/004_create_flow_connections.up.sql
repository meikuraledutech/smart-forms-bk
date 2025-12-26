CREATE TABLE IF NOT EXISTS flow_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    question_id UUID NOT NULL REFERENCES questions(id) ON DELETE RESTRICT,
    parent_id UUID REFERENCES flow_connections(id) ON DELETE CASCADE,
    order_index INT NOT NULL DEFAULT 0,
    depth_level INT NOT NULL DEFAULT 0,
    is_terminal BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_flow_connections_form_id ON flow_connections(form_id);
CREATE INDEX IF NOT EXISTS idx_flow_connections_question_id ON flow_connections(question_id);
CREATE INDEX IF NOT EXISTS idx_flow_connections_parent_id ON flow_connections(parent_id);
CREATE INDEX IF NOT EXISTS idx_flow_connections_deleted_at ON flow_connections(deleted_at);
