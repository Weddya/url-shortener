-- +goose Up
CREATE TABLE urls (
                      id SERIAL PRIMARY KEY,
                      original_url TEXT NOT NULL UNIQUE,
                      short_code VARCHAR(10) NOT NULL UNIQUE,
                      created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_short_code ON urls (short_code);

-- +goose Down
DROP TABLE urls;