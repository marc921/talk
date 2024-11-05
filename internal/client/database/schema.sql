CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE local_users (
	name TEXT PRIMARY KEY,
	private_key BLOB
);
CREATE TABLE public_users (
	name TEXT PRIMARY KEY,
	public_key BLOB
);
CREATE TABLE conversations (
	id INTEGER PRIMARY KEY,
	local_user_name TEXT REFERENCES local_users(name) NOT NULL,
	remote_user_name TEXT REFERENCES public_users(name) NOT NULL,
	last_symmetric_key BLOB,
	UNIQUE (local_user_name, remote_user_name)
);
CREATE TABLE messages (
	id INTEGER PRIMARY KEY,
	conversation_id INT REFERENCES conversations(id) NOT NULL,
	sender TEXT REFERENCES public_users(name) NOT NULL,
	receiver TEXT REFERENCES public_users(name) NOT NULL,
	content BLOB,
	sent_at DATETIME,
	delivered_at DATETIME,
	read_at DATETIME
);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20241105135553'),
  ('20241105140048'),
  ('20241105142502');
