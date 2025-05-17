-- Bank Accounts Table
CREATE TABLE bank_accounts (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    account_id INT NOT NULL REFERENCES accounts(id),
    account_number VARCHAR(255) NOT NULL,
    bank_name VARCHAR(255) NOT NULL,
    branch_name VARCHAR(255),
    swift_code VARCHAR(50),
    UNIQUE (account_id)
);

-- Bank Accounts Table Indexes
CREATE INDEX idx_bank_accounts_account_id ON bank_accounts(account_id);
CREATE INDEX idx_bank_accounts_deleted_at ON bank_accounts(deleted_at);