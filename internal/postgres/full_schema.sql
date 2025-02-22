-- full_schema.sql

CREATE TABLE users (
    id UUID PRIMARY KEY,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE isas (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    fund_ids UUID[] NOT NULL,
    cash_balance DECIMAL(15,2) DEFAULT 0,
    investment_amount DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);


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

CREATE TABLE investments (
    id UUID PRIMARY KEY,
    isa_id UUID REFERENCES isas(id) ON DELETE SET NULL,
    fund_id UUID REFERENCES funds(id) ON DELETE SET NULL,
    amount DECIMAL(15,2) NOT NULL,
    invested_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
