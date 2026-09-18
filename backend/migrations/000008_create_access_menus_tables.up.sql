-- Create access_menus and access_menus_per_roles tables (UP)
CREATE TABLE IF NOT EXISTS access_menus (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(150) NOT NULL,
    path VARCHAR(255) NOT NULL,
    icon VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS access_menus_per_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role user_role NOT NULL,
    menu_key VARCHAR(100) NOT NULL REFERENCES access_menus(key) ON DELETE CASCADE,
    can_access BOOLEAN NOT NULL DEFAULT true,
    UNIQUE(role, menu_key)
);

CREATE INDEX IF NOT EXISTS idx_access_menus_role ON access_menus_per_roles(role);
