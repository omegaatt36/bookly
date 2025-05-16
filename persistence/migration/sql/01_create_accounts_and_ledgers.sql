-- Create accounts table
CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
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

-- Create ledgers table
CREATE TABLE ledgers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    account_id INT NOT NULL REFERENCES accounts(id),
    date TIMESTAMP WITH TIME ZONE NOT NULL,
    type VARCHAR(20) NOT NULL,
    amount DECIMAL(20, 2) NOT NULL,
    note TEXT,
    is_adjustment BOOLEAN NOT NULL DEFAULT FALSE,
    adjusted_from INT REFERENCES ledgers(id),
    is_voided BOOLEAN NOT NULL DEFAULT FALSE,
    voided_at TIMESTAMP WITH TIME ZONE
);

-- Ledgers Table Indexes
CREATE INDEX idx_ledgers_account_id ON ledgers (account_id);
CREATE INDEX idx_ledgers_date ON ledgers (date);
CREATE INDEX idx_ledgers_deleted_at ON ledgers (deleted_at);
CREATE INDEX idx_ledgers_adjusted_from ON ledgers (adjusted_from);