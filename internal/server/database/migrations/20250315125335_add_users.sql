-- migrate:up
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

-- migrate:down
DROP TABLE users;
DROP TABLE messages;
