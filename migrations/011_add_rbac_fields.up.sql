-- Add role column to users table
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user';
CREATE INDEX idx_users_role ON users(role);

-- Add template fields to forms table
ALTER TABLE forms ADD COLUMN is_template BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE forms ADD COLUMN template_category TEXT;
ALTER TABLE forms ADD COLUMN template_description TEXT;
CREATE INDEX idx_forms_is_template ON forms(is_template) WHERE is_template = true;

-- Create subscription_plans table (managed by super admin)
CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    plan_type TEXT NOT NULL,
    price_inr INTEGER NOT NULL DEFAULT 0,
    razorpay_plan_id TEXT UNIQUE,
    features JSONB DEFAULT '{}'::jsonb,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

-- Seed default free plan
INSERT INTO subscription_plans (name, plan_type, price_inr, features)
VALUES ('Free', 'free', 0, '{"data_retention_days": 7, "can_export": false}');
