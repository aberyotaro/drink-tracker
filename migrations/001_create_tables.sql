-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slack_user_id TEXT UNIQUE NOT NULL,
    slack_team_id TEXT NOT NULL,
    daily_limit_ml INTEGER DEFAULT 40000,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create drink_types table
CREATE TABLE IF NOT EXISTS drink_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    typical_amount_ml INTEGER NOT NULL,
    alcohol_percentage REAL NOT NULL
);

-- Create drink_records table
CREATE TABLE IF NOT EXISTS drink_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    drink_type TEXT NOT NULL,
    amount_ml INTEGER NOT NULL,
    alcohol_percentage REAL NOT NULL,
    recorded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- Insert default drink types (デフォルトは350ml、500mlは引数で指定)
INSERT OR IGNORE INTO drink_types (name, typical_amount_ml, alcohol_percentage) VALUES
('beer', 350, 0.05),
('wine', 150, 0.12),
('sake', 180, 0.15),
('whiskey', 30, 0.40),
('shochu', 60, 0.25),
('highball', 350, 0.09);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_slack_user_id ON users(slack_user_id);
CREATE INDEX IF NOT EXISTS idx_drink_records_user_id ON drink_records(user_id);
CREATE INDEX IF NOT EXISTS idx_drink_records_recorded_at ON drink_records(recorded_at);