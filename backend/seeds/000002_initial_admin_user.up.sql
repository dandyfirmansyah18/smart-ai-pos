-- Initial Admin User Seed (UP)
INSERT INTO users (id, username, password_hash, role, full_name, created_at, updated_at)
VALUES (
    'b0000000-0000-0000-0000-000000000001',
    'admin',
    '$2a$10$0H0/uR2z21WPyfhlGtvptOaoNQJiUcMsRVG6gHAyg4Dehfn3T.f8G',
    'ADMIN',
    'System Administrator',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
)
ON CONFLICT (username) DO UPDATE SET
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    full_name = EXCLUDED.full_name,
    updated_at = CURRENT_TIMESTAMP;
