CREATE TABLE isas (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    fund_ids UUID[] DEFAULT NULL,
    cash_balance DECIMAL(15,2) DEFAULT 0,
    investment_amount DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);