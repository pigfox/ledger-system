CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE user_addresses (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id),
    chain TEXT NOT NULL,
    address TEXT NOT NULL,
    UNIQUE (user_id, chain, address)
);

CREATE TABLE transactions (
    id UUID PRIMARY KEY,
    user_id INT REFERENCES users(id),
    type TEXT NOT NULL, -- deposit, withdrawal, transfer
    amount NUMERIC NOT NULL,
    currency TEXT NOT NULL, -- ETH, MATIC, USDC
    status TEXT NOT NULL,
    tx_hash TEXT,
    block_height INT,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE ledger_entries (
    id UUID PRIMARY KEY,
    transaction_id UUID REFERENCES transactions(id),
    account TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    currency TEXT NOT NULL,
    direction TEXT NOT NULL, -- credit / debit
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE onchain_transactions (
    id UUID PRIMARY KEY,
    address TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    amount NUMERIC NOT NULL,
    currency TEXT NOT NULL,
    direction TEXT NOT NULL,
    block_height INT NOT NULL,
    confirmed BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (tx_hash, address)
);