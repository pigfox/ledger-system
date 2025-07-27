#!/bin/bash
set -euo pipefail

DB_URL="postgres://xyz:xyzpass@localhost:5432/xyzledger"
API_URL="http://localhost:8080/api/v1/reconciliation"
ADDRESS="0xdadb0d80178819f2319190d340ce9a924f783711"

echo "Inserting invalid on-chain transactions..."

psql "$DB_URL" <<EOF
-- TX with negative amount
INSERT INTO onchain_transactions (
    id, address, tx_hash, amount, currency, direction, block_height, reconciled
) VALUES (
    gen_random_uuid(),
    '$ADDRESS',
    '0xaaaabbbbccccddddeeeeffff0000111122223333444455556666777788889999',
    -42.0,
    'ETH',
    'credit',
    12345679,
    false
);

-- TX with unknown address
INSERT INTO onchain_transactions (
    id, address, tx_hash, amount, currency, direction, block_height, reconciled
) VALUES (
    gen_random_uuid(),
    '0x000000000000000000000000000000000000dEaD',
    '0xffffeeee111122223333444455556666777788889999aaaabbbbccccddddeeee',
    1.23,
    'ETH',
    'debit',
    12345680,
    false
);

-- TX with invalid currency
INSERT INTO onchain_transactions (
    id, address, tx_hash, amount, currency, direction, block_height, reconciled
) VALUES (
    gen_random_uuid(),
    '$ADDRESS',
    '0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef',
    5.0,
    'DOGE',
    'credit',
    12345681,
    false
);

-- TX with zero amount
INSERT INTO onchain_transactions (
    id, address, tx_hash, amount, currency, direction, block_height, reconciled
) VALUES (
    gen_random_uuid(),
    '$ADDRESS',
    '0xeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee',
    0.0,
    'ETH',
    'credit',
    12345682,
    false
);
EOF

echo "Inserted invalid transactions."
