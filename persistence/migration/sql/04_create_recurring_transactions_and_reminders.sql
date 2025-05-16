-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Recurring Transactions Table
CREATE TABLE recurring_transactions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INT NOT NULL REFERENCES users(id),
    account_id INT NOT NULL REFERENCES accounts(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    note TEXT,
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE,
    recur_type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    frequency INT NOT NULL DEFAULT 1,
    day_of_week INT,
    day_of_month INT,
    month_of_year INT,
    last_executed TIMESTAMP WITH TIME ZONE,
    next_due TIMESTAMP WITH TIME ZONE NOT NULL
);

-- Reminders Table
CREATE TABLE reminders (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    recurring_transaction_id INT NOT NULL REFERENCES recurring_transactions(id),
    reminder_date TIMESTAMP WITH TIME ZONE NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE
);

-- Index for faster lookups
CREATE INDEX idx_recurring_transactions_user_id ON recurring_transactions(user_id);
CREATE INDEX idx_recurring_transactions_next_due ON recurring_transactions(next_due);
CREATE INDEX idx_recurring_transactions_status ON recurring_transactions(status);
CREATE INDEX idx_reminders_recurring_transaction_id ON reminders(recurring_transaction_id);
CREATE INDEX idx_reminders_reminder_date ON reminders(reminder_date);
CREATE INDEX idx_reminders_is_read ON reminders(is_read);