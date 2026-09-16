-- Down migration for adding kitchen order status ENUM values
-- PostgreSQL does not support dropping ENUM values directly via DROP TYPE without recreation.
SELECT 1;
