
-- Accounts Table
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE,
        user_id INT NOT NULL,
        name VARCHAR(255) NOT NULL,
        status VARCHAR(20) NOT NULL,
        currency VARCHAR(3) NOT NULL,
        balance DECIMAL(20, 2) NOT NULL
);

-- Accounts Table Indexes
CREATE INDEX idx_accounts_user_id ON accounts (user_id);

CREATE INDEX idx_accounts_deleted_at ON accounts (deleted_at);

CREATE INDEX idx_accounts_status ON accounts (status);

-- Ledgers Table
CREATE TABLE ledgers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE,
        account_id INT NOT NULL REFERENCES accounts (id),
        date TIMESTAMP
    WITH
        TIME ZONE NOT NULL,
        type VARCHAR(20) NOT NULL,
        amount DECIMAL(20, 2) NOT NULL,
        note TEXT,
        category_id INT REFERENCES categories (id), -- Can be NULL if a ledger item doesn't belong to a category
        is_adjustment BOOLEAN NOT NULL DEFAULT FALSE,
        adjusted_from INT REFERENCES ledgers (id),
        is_voided BOOLEAN NOT NULL DEFAULT FALSE,
        voided_at TIMESTAMP
    WITH
        TIME ZONE
);

-- Ledgers Table Indexes
CREATE INDEX idx_ledgers_account_id ON ledgers (account_id);

CREATE INDEX idx_ledgers_date ON ledgers (date);

CREATE INDEX idx_ledgers_deleted_at ON ledgers (deleted_at);

CREATE INDEX idx_ledgers_adjusted_from ON ledgers (adjusted_from);

CREATE INDEX idx_ledgers_category_id ON ledgers (category_id);

-- Users Table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE,
        disabled BOOLEAN NOT NULL DEFAULT FALSE,
        name VARCHAR(255) NOT NULL,
        nickname VARCHAR(255)
);

-- Users Table Indexes
CREATE INDEX idx_users_deleted_at ON users (deleted_at);

CREATE INDEX idx_users_disabled ON users (disabled);

-- Identities Table
CREATE TABLE identities (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users (id),
    provider VARCHAR(20) NOT NULL,
    identifier VARCHAR(255) NOT NULL,
    credential VARCHAR(255) NOT NULL,
    last_used_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        UNIQUE (user_id, provider, identifier),
        UNIQUE (provider, identifier)
);

-- Identities Table Indexes
CREATE INDEX idx_identities_user_id ON identities (user_id);

CREATE INDEX idx_identities_last_used_at ON identities (last_used_at);

-- Categories Table
CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INT NOT NULL REFERENCES users (id), -- Each user has their own categories
    name VARCHAR(255) NOT NULL,
    UNIQUE (user_id, name) -- Category names must be unique per user
);

CREATE INDEX idx_categories_user_id ON categories (user_id);
CREATE INDEX idx_categories_deleted_at ON categories (deleted_at);

-- Budgets Table
CREATE TABLE budgets (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INT NOT NULL REFERENCES users (id),
    name VARCHAR(255) NOT NULL,
    period VARCHAR(20) NOT NULL, -- 'monthly' or 'yearly'
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL, -- Calculated based on start_date and period
    amount DECIMAL(20, 2) NOT NULL,
    category_id INT NOT NULL REFERENCES categories (id)
);

-- Budgets Table Indexes
CREATE INDEX idx_budgets_user_id ON budgets (user_id);
CREATE INDEX idx_budgets_category_id ON budgets (category_id);
CREATE INDEX idx_budgets_deleted_at ON budgets (deleted_at);
CREATE INDEX idx_budgets_period ON budgets (period);
CREATE INDEX idx_budgets_start_date ON budgets (start_date);
CREATE INDEX idx_budgets_end_date ON budgets (end_date);

-- Recurring Transactions Table
CREATE TABLE recurring_transactions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE,
        user_id INT NOT NULL REFERENCES users (id),
        account_id INT NOT NULL REFERENCES accounts (id),
        name VARCHAR(255) NOT NULL,
        type VARCHAR(20) NOT NULL,
        amount DECIMAL(20, 2) NOT NULL,
        note TEXT,
        start_date TIMESTAMP
    WITH
        TIME ZONE NOT NULL,
        end_date TIMESTAMP
    WITH
        TIME ZONE,
        recur_type VARCHAR(20) NOT NULL,
        status VARCHAR(20) NOT NULL DEFAULT 'active',
        frequency INT NOT NULL DEFAULT 1,
        day_of_week INT,
        day_of_month INT,
        month_of_year INT,
        last_executed TIMESTAMP
    WITH
        TIME ZONE,
        next_due TIMESTAMP
    WITH
        TIME ZONE NOT NULL
);

-- Reminders Table
CREATE TABLE reminders (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE,
        recurring_transaction_id INT NOT NULL REFERENCES recurring_transactions (id),
        reminder_date TIMESTAMP
    WITH
        TIME ZONE NOT NULL,
        is_read BOOLEAN NOT NULL DEFAULT FALSE,
        read_at TIMESTAMP
    WITH
        TIME ZONE
);

-- Index for faster lookups
CREATE INDEX idx_recurring_transactions_user_id ON recurring_transactions (user_id);

CREATE INDEX idx_recurring_transactions_next_due ON recurring_transactions (next_due);

CREATE INDEX idx_recurring_transactions_status ON recurring_transactions (status);

-- Reminders Table Indexes
CREATE INDEX idx_reminders_recurring_transaction_id ON reminders (recurring_transaction_id);

CREATE INDEX idx_reminders_reminder_date ON reminders (reminder_date);

CREATE INDEX idx_reminders_is_read ON reminders (is_read);

-- Bank Accounts Table
CREATE TABLE bank_accounts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        updated_at TIMESTAMP
    WITH
        TIME ZONE NOT NULL DEFAULT NOW (),
        deleted_at TIMESTAMP
    WITH
        TIME ZONE,
        account_id INT NOT NULL REFERENCES accounts (id),
        account_number VARCHAR(255) NOT NULL,
        bank_name VARCHAR(255) NOT NULL,
        branch_name VARCHAR(255),
        swift_code VARCHAR(50),
        UNIQUE (account_id)
);

-- Bank Accounts Table Indexes
CREATE INDEX idx_bank_accounts_account_id ON bank_accounts (account_id);
CREATE INDEX idx_bank_accounts_deleted_at ON bank_accounts (deleted_at);
