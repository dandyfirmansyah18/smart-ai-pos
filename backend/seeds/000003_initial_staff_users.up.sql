-- Initial Staff Users Seed (Cashier, Kitchen, Warehouse) (UP)
INSERT INTO users (id, username, password_hash, role, full_name, created_at, updated_at)
VALUES 
    (
        'b0000000-0000-0000-0000-000000000002'::uuid,
        'cashier',
        '$2a$10$fxHpOYCmOLOdAmcht1zgqe6GU6S/gnPYiyEI7jvzrijRVbqKKpKm6',
        'CASHIER',
        'Front Cashier Staff',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'b0000000-0000-0000-0000-000000000003'::uuid,
        'kitchen',
        '$2a$10$txEctljEGpbBxbsq8RYvqOOtX7W9Qw2AKUmro19HCl00xyeQLmt7y',
        'KITCHEN',
        'Head Kitchen Chef',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    ),
    (
        'b0000000-0000-0000-0000-000000000004'::uuid,
        'warehouse',
        '$2a$10$rTdlw7RirDkDTRu789721O/BiXUUhckIP8WDPvQDRWY9vi0vx2i9K',
        'WAREHOUSE',
        'Inventory Warehouse Manager',
        CURRENT_TIMESTAMP,
        CURRENT_TIMESTAMP
    )
ON CONFLICT (username) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    full_name = EXCLUDED.full_name,
    updated_at = CURRENT_TIMESTAMP;
