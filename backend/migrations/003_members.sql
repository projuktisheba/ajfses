-- Member's teams
CREATE TABLE teams (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP

);
CREATE INDEX idx_teams_title ON teams(title);

-- Insert data into 'teams'
INSERT INTO teams (title)
VALUES
('Management'),
('Admin, HR, Accounts & Commercial'),
('Business Development Team'),
('Design Team'),
('Engineering Team'),
('Service & Technical Team');

CREATE TABLE members (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    contact VARCHAR(255) NOT NULL DEFAULT '',
    team BIGINT DEFAULT 0,
    designation TEXT DEFAULT '',
    note TEXT DEFAULT '',
    image_link TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes 

CREATE INDEX idx_members_name ON members(name);
CREATE INDEX idx_members_contact ON members(contact);
CREATE INDEX idx_members_team ON members(team);
CREATE INDEX idx_members_designation ON members(designation);