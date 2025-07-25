CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL CHECK (LENGTH(name) >= 2),
    email VARCHAR(50) NOT NULL UNIQUE CHECK (LENGTH(email) <= 50),
    password TEXT NOT NULL CHECK (LENGTH(password) >= 5),
    city VARCHAR(100),
    state VARCHAR(100),
    country VARCHAR(100),
    occupation VARCHAR(100),
    phone_number VARCHAR(20),
    transactions UUID[] DEFAULT '{}'::UUID[],
    role VARCHAR(20) NOT NULL DEFAULT 'admin' 
        CHECK (role IN ('user', 'admin', 'superadmin')),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
