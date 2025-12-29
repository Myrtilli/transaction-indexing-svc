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

CREATE TABLE IF NOT EXISTS block_headers (
    block_hash     text PRIMARY KEY,
    previous_hash  text NOT NULL,
    transaction_num bigint NOT NULL,
    height         bigint NOT NULL UNIQUE,
    merkle_root    text NOT NULL,
    timestamp      timestamp NOT NULL,
    difficulty     bigint NOT NULL,
    nonce          bigint NOT NULL
);

CREATE TABLE IF NOT EXISTS transactions (
    id            bigserial PRIMARY KEY,
    tx_id         text NOT NULL,
    address_id    bigint NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
    amount        bigint NOT NULL,
    block_height  bigint NOT NULL REFERENCES block_headers(height) ON DELETE CASCADE,
    block_hash    text NOT NULL REFERENCES block_headers(block_hash) ON DELETE CASCADE,
    merkle_proof  text[] NOT NULL,
    created_at    timestamp DEFAULT now()
);

CREATE TABLE IF NOT EXISTS utxos (
    id            bigserial PRIMARY KEY,
    address_id    bigint NOT NULL REFERENCES addresses(id) ON DELETE CASCADE,
    tx_id         text NOT NULL,
    vout          int NOT NULL,
    amount        bigint NOT NULL,
    block_height  bigint NOT NULL REFERENCES block_headers(height) ON DELETE CASCADE,
    is_spent      boolean DEFAULT false,
    UNIQUE(tx_id, vout)
);

CREATE INDEX IF NOT EXISTS idx_addresses_user_id ON addresses(user_id);
CREATE INDEX IF NOT EXISTS idx_utxos_address_id ON utxos(address_id) WHERE is_spent = false;
CREATE INDEX IF NOT EXISTS idx_transactions_address_id ON transactions(address_id);

-- +migrate Down
DROP TABLE IF EXISTS utxos;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS block_headers;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS users;