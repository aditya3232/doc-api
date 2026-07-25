CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    phone varchar(17) NULL,
    photo varchar(255) NULL,
    address TEXT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Unique hanya untuk data yang belum di-soft delete
CREATE UNIQUE INDEX idx_users_email_unique
ON users (email)
WHERE deleted_at IS NULL;