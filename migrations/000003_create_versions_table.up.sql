CREATE TABLE versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    package_id UUID REFERENCES packages(id) ON DELETE CASCADE,
    version TEXT NOT NULL,
    file_key TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(package_id, version)
);
