-- Analytics Status Table
CREATE TABLE IF NOT EXISTS analytics_status (
    form_id UUID PRIMARY KEY REFERENCES forms(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'calculating', 'completed', 'failed')),
    calculated_at TIMESTAMPTZ,
    triggered_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

-- Index for status queries
CREATE INDEX idx_analytics_status_form_id ON analytics_status(form_id);
CREATE INDEX idx_analytics_status_status ON analytics_status(status);

-- Node Analytics Table
CREATE TABLE IF NOT EXISTS analytics_nodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    flow_connection_id UUID NOT NULL REFERENCES flow_connections(id) ON DELETE CASCADE,
    visit_count INT NOT NULL DEFAULT 0,
    answer_count INT NOT NULL DEFAULT 0,
    skip_count INT NOT NULL DEFAULT 0,
    drop_off_count INT NOT NULL DEFAULT 0,
    total_time_spent INT NOT NULL DEFAULT 0,
    avg_time_spent FLOAT NOT NULL DEFAULT 0,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    UNIQUE(form_id, flow_connection_id)
);

-- Indexes for node analytics queries
CREATE INDEX idx_analytics_nodes_form_id ON analytics_nodes(form_id);
CREATE INDEX idx_analytics_nodes_flow_connection_id ON analytics_nodes(flow_connection_id);
CREATE INDEX idx_analytics_nodes_visit_count ON analytics_nodes(visit_count DESC);
CREATE INDEX idx_analytics_nodes_avg_time_spent ON analytics_nodes(avg_time_spent DESC);

-- Path Analytics Table
CREATE TABLE IF NOT EXISTS analytics_paths (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    form_id UUID NOT NULL REFERENCES forms(id) ON DELETE CASCADE,
    path JSONB NOT NULL,
    occurrence_count INT NOT NULL DEFAULT 0,
    avg_completion_time FLOAT NOT NULL DEFAULT 0,
    completion_rate FLOAT NOT NULL DEFAULT 0,
    calculated_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

-- Indexes for path analytics queries
CREATE INDEX idx_analytics_paths_form_id ON analytics_paths(form_id);
CREATE INDEX idx_analytics_paths_occurrence_count ON analytics_paths(occurrence_count DESC);
CREATE INDEX idx_analytics_paths_path ON analytics_paths USING GIN (path);

-- Comments
COMMENT ON TABLE analytics_status IS 'Tracks the status of analytics calculation for each form';
COMMENT ON TABLE analytics_nodes IS 'Stores node-level analytics metrics';
COMMENT ON TABLE analytics_paths IS 'Stores path-level analytics metrics';
