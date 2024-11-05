-- migrate:up
CREATE TABLE conversations (
	id INTEGER PRIMARY KEY,
	local_user_name TEXT REFERENCES local_users(name) NOT NULL,
	remote_user_name TEXT REFERENCES public_users(name) NOT NULL,
	last_symmetric_key BLOB,
	UNIQUE (local_user_name, remote_user_name)
);

-- migrate:down
DROP TABLE conversations;
