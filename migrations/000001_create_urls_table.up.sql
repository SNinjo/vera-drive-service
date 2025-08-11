CREATE TABLE url_nodes (
  id UUID PRIMARY KEY,
  user_id INTEGER NOT NULL,
  parent_id UUID,
  name VARCHAR(255) NOT NULL,
  type VARCHAR(10) NOT NULL CHECK (type IN ('folder', 'url')),
  url TEXT,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL,
  deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_url_nodes_user_id ON url_nodes(user_id);
CREATE INDEX idx_url_nodes_parent_id ON url_nodes(parent_id);
CREATE INDEX idx_url_nodes_deleted_at ON url_nodes(deleted_at);

ALTER TABLE url_nodes ADD CONSTRAINT fk_url_nodes_parent 
  FOREIGN KEY (parent_id) REFERENCES url_nodes(id);
