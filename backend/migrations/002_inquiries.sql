CREATE TABLE inquiries (
    id BIGSERIAL PRIMARY KEY,

    inquiry_date TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,

    name VARCHAR(100) NOT NULL,
    mobile VARCHAR(255) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL DEFAULT '',
    subject TEXT DEFAULT '',
    message TEXT DEFAULT '',

    status VARCHAR(20) NOT NULL DEFAULT 'NEW' CHECK (
        status IN (
            'NEW',           -- just arrived
            'IN PROGRESS',   -- being processed
            'RESOLVED'     -- issue solved
        )
    ),

    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP

);

-- Indexes 
CREATE INDEX idx_inquiries_inquiry_date ON inquiries(inquiry_date);
CREATE INDEX idx_inquiries_name ON inquiries(name);
CREATE INDEX idx_inquiries_mobile ON inquiries(mobile);
CREATE INDEX idx_inquiries_email ON inquiries(email);
CREATE INDEX idx_inquiries_status ON inquiries(status);