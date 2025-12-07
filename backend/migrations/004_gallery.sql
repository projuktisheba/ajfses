-- Gallery's table
CREATE TABLE gallery (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(100) NOT NULL,
    image_link TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP

);
CREATE INDEX idx_gallery_title ON gallery(title);