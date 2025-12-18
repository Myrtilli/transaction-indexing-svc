-- +migrate Up
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    hashed_password TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS addresses (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    address TEXT NOT NULL,
    UNIQUE(user_id, address)
);

CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses(user_id);

-- +migrate Down
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS addresses;
DROP INDEX IF EXISTS idx_addresses_user_id;