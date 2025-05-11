-- Create users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    disabled BOOLEAN NOT NULL DEFAULT FALSE,
    name VARCHAR(255) NOT NULL,
    nickname VARCHAR(255)
);

-- Users Table Indexes
CREATE INDEX idx_users_deleted_at ON users (deleted_at);
CREATE INDEX idx_users_disabled ON users (disabled);

-- Add user_id column to accounts if it doesn't exist already
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'accounts' AND column_name = 'user_id'
    ) THEN
        ALTER TABLE accounts ADD COLUMN user_id UUID;
    END IF;
END $$;

-- Find or create default user
DO $$
DECLARE
    default_user_id UUID;
BEGIN
    -- Check if we have accounts with null user_id
    IF EXISTS (SELECT 1 FROM accounts WHERE user_id IS NULL) THEN
        -- Create default user if it doesn't exist
        INSERT INTO users (name, nickname) 
        VALUES ('default', 'default')
        RETURNING id INTO default_user_id;
        
        -- Update all accounts with null user_id
        UPDATE accounts SET user_id = default_user_id WHERE user_id IS NULL;
    END IF;
END $$;

-- Make user_id column not null if not already
ALTER TABLE accounts ALTER COLUMN user_id SET NOT NULL;