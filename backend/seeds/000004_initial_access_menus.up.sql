-- Initial Access Menus Master & Role Mappings (UP)
INSERT INTO access_menus (key, name, path, icon)
VALUES 
    ('POS', 'POS Terminal', '/checkout', 'ShoppingBag'),
    ('KDS', 'Kitchen (KDS)', '/portal/kitchen', 'ChefHat'),
    ('WAREHOUSE', 'Warehouse', '/portal/warehouse', 'Warehouse'),
    ('FINANCE', 'Finance P&L', '/portal/finance', 'TrendingUp'),
    ('ANALYTICS', 'Analytics', '/dashboard', 'BarChart3'),
    ('RECEIPT', 'Receipt Scanner', '/receipts', 'Receipt'),
    ('SETUP_ROLE', 'Setup Role', '/portal/setup-role', 'ShieldCheck')
ON CONFLICT (key) DO UPDATE SET name = EXCLUDED.name, path = EXCLUDED.path, icon = EXCLUDED.icon;

-- Default access mappings per role
INSERT INTO access_menus_per_roles (role, menu_key, can_access)
VALUES 
    -- ADMIN
    ('ADMIN', 'POS', true),
    ('ADMIN', 'KDS', true),
    ('ADMIN', 'WAREHOUSE', true),
    ('ADMIN', 'FINANCE', true),
    ('ADMIN', 'ANALYTICS', true),
    ('ADMIN', 'RECEIPT', true),
    ('ADMIN', 'SETUP_ROLE', true),
    
    -- CASHIER
    ('CASHIER', 'POS', true),
    ('CASHIER', 'KDS', true),
    ('CASHIER', 'WAREHOUSE', false),
    ('CASHIER', 'FINANCE', true),
    ('CASHIER', 'ANALYTICS', true),
    ('CASHIER', 'RECEIPT', true),
    ('CASHIER', 'SETUP_ROLE', false),

    -- KITCHEN
    ('KITCHEN', 'POS', false),
    ('KITCHEN', 'KDS', true),
    ('KITCHEN', 'WAREHOUSE', false),
    ('KITCHEN', 'FINANCE', false),
    ('KITCHEN', 'ANALYTICS', false),
    ('KITCHEN', 'RECEIPT', false),
    ('KITCHEN', 'SETUP_ROLE', false),

    -- WAREHOUSE
    ('WAREHOUSE', 'POS', false),
    ('WAREHOUSE', 'KDS', false),
    ('WAREHOUSE', 'WAREHOUSE', true),
    ('WAREHOUSE', 'FINANCE', false),
    ('WAREHOUSE', 'ANALYTICS', false),
    ('WAREHOUSE', 'RECEIPT', false),
    ('WAREHOUSE', 'SETUP_ROLE', false)
ON CONFLICT (role, menu_key) DO UPDATE SET can_access = EXCLUDED.can_access;
