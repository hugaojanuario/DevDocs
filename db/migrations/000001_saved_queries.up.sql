CREATE TABLE saved_queries (
 id UUID PRIMARY KEY,
 question TEXT NOT NULL,
 title VARCHAR(255) NOT NULL,
 technology VARCHAR(100),
 response JSONB NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX saved_queries_created_idx ON saved_queries (created_at DESC, id DESC);
CREATE INDEX saved_queries_technology_idx ON saved_queries (technology);
