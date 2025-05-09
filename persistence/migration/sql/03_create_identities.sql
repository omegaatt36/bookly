-- Create identities table
CREATE TABLE identities (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    provider VARCHAR(20) NOT NULL,
    identifier VARCHAR(255) NOT NULL,
    credential VARCHAR(255) NOT NULL,
    last_used_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_user_provider_identifier UNIQUE (user_id, provider, identifier),
    CONSTRAINT unique_provider_identifier UNIQUE (provider, identifier)
);

-- Add indexes for performance
CREATE INDEX idx_identities_user_id ON identities(user_id);
CREATE INDEX idx_identities_provider_identifier ON identities(provider, identifier);