-- Rollback Staff Users Seed (DOWN)
DELETE FROM users WHERE username IN ('cashier', 'kitchen', 'warehouse');
