CREATE TABLE IF NOT EXISTS credentials (
    id TEXT PRIMARY KEY,
    profile_id TEXT NOT NULL,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    token TEXT,
    username TEXT,
    password TEXT,
    header_name TEXT NOT NULL,
    header_value TEXT NOT NULL,
    
    -- OAuth2 Specific Fields
    client_id TEXT,
    client_secret TEXT,
    redirect_uri TEXT,
    refresh_token TEXT,
    token_expiry TIMESTAMP,
    
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (profile_id) REFERENCES profiles(id) ON DELETE CASCADE
); 