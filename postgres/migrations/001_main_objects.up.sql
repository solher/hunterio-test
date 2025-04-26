SET SCHEMA 'hunterio';
BEGIN;
----

CREATE TABLE extracted_data (
  id SERIAL PRIMARY KEY,
  url TEXT NOT NULL,
  people JSONB NOT NULL DEFAULT '{}',
  companies JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX extracted_data_by_url ON extracted_data (url, created_at);

----
COMMIT;
