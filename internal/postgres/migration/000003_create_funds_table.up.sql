CREATE TABLE funds (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    type VARCHAR(50) CHECK (type IN ('Equity', 'Bond', 'Index', 'Mixed')),
    risk_level VARCHAR(50) CHECK (risk_level IN ('Low', 'Medium', 'High')),
    performance DECIMAL(15,2) DEFAULT 0,
    total_amount DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);