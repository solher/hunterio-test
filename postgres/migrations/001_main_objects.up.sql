SET SCHEMA 'hunterio';
BEGIN;
----

CREATE FUNCTION set_updated_at_column()
  RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

--

CREATE TABLE extracted_data (
  id SERIAL PRIMARY KEY,
  url TEXT UNIQUE NOT NULL,
  people JSONB NOT NULL DEFAULT '{}',
  companies JSONB NOT NULL DEFAULT '{}',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TRIGGER set_updated_at_extracted_data
BEFORE UPDATE ON extracted_data
FOR EACH ROW
EXECUTE PROCEDURE set_updated_at_column();

----
COMMIT;
