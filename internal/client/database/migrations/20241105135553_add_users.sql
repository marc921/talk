-- migrate:up
CREATE TABLE local_users (
	name TEXT PRIMARY KEY,
	private_key BLOB
);

CREATE TABLE public_users (
	name TEXT PRIMARY KEY,
	public_key BLOB
);

-- migrate:down
DROP TABLE users;
DROP TABLE public_users;
