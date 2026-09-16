-- Initial Products Catalog Seed (DOWN / Rollback)
DELETE FROM products WHERE sku IN ('SKU-COFFEE-001', 'SKU-COFFEE-002', 'SKU-FOOD-001', 'SKU-FOOD-002', 'SKU-DRINK-001');
