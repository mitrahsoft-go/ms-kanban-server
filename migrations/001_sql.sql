-- Organizations

CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    domain VARCHAR(150) NOT NULL UNIQUE,
    logo_url VARCHAR(150) NOT NULL,

    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_organizations_deleted_at
ON organizations (deleted_at);


-- Users

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    organization_id UUID,
    full_name VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT,
    role VARCHAR(30),
    avatar_url VARCHAR(255),
    timezone VARCHAR(50) DEFAULT 'UTC',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    
    CONSTRAINT fk_users_organization
        FOREIGN KEY (organization_id)
        REFERENCES organizations(id)
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at
ON users (deleted_at);

CREATE INDEX IF NOT EXISTS idx_users_organization_id
ON users (organization_id);

CREATE INDEX IF NOT EXISTS idx_users_email
ON users (email);

CREATE INDEX IF NOT EXISTS idx_users_role
ON users (role);