-- migrate:up
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

-- migrate:down
DROP TABLE messages;
