-- Database initialization script

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    balance NUMERIC(15, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    idempotency_key VARCHAR UNIQUE NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    balance_before NUMERIC(15, 2) NOT NULL,
    balance_after NUMERIC(15, 2) NOT NULL,
    amount NUMERIC(15, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

INSERT INTO users (id, balance)
VALUES (1, 100.00)
ON CONFLICT (id) DO NOTHING;

-- Update sequence to start from 2 for future users
SELECT setval('users_id_seq', (SELECT MAX(id) FROM users));
