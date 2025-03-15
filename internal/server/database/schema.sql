CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE users (
	name TEXT PRIMARY KEY,
	public_key BLOB NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE messages (
	id INTEGER PRIMARY KEY,
	sender TEXT references users(name) NOT NULL,
	recipient TEXT references users(name) NOT NULL,
	cipher_sym_key BLOB NOT NULL,
	ciphertext BLOB NOT NULL,
	sent_at TIMESTAMP,
	delivered_at TIMESTAMP,
	read_at TIMESTAMP
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20250315125335');
