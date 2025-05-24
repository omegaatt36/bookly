-- Add category column to ledgers table
ALTER TABLE ledgers ADD COLUMN category VARCHAR(100);

-- Create index for category
CREATE INDEX idx_ledgers_category ON ledgers (category);

-- Create budgets table
CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INT NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    period_type VARCHAR(20) NOT NULL, -- 'monthly' or 'yearly'
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Budgets Table Indexes
CREATE INDEX idx_budgets_user_id ON budgets (user_id);
CREATE INDEX idx_budgets_deleted_at ON budgets (deleted_at);
CREATE INDEX idx_budgets_category ON budgets (category);
CREATE INDEX idx_budgets_period_type ON budgets (period_type);
CREATE INDEX idx_budgets_start_date ON budgets (start_date);
CREATE INDEX idx_budgets_end_date ON budgets (end_date);
CREATE INDEX idx_budgets_is_active ON budgets (is_active);