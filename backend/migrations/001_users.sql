-- =========================
-- Table: users
-- =========================
-- Create users table
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'Operator',
    status VARCHAR(20) NOT NULL DEFAULT 'Active' CHECK (status IN ('Active', 'Inactive')),
    mobile VARCHAR(20) NOT NULL,
    email VARCHAR(150) NOT NULL DEFAULT '',
    password TEXT NOT NULL DEFAULT '',
    joining_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    address VARCHAR(1000) NOT NULL DEFAULT '',
    avatar_link TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes 
CREATE INDEX idx_users_name ON users(name);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);
CREATE INDEX idx_users_mobile ON users(mobile);
CREATE INDEX idx_users_email ON users(email);
