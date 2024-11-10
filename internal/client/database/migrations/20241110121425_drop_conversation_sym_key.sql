-- migrate:up
ALTER TABLE conversations DROP COLUMN last_symmetric_key;

-- migrate:down
ALTER TABLE conversations ADD COLUMN last_symmetric_key BLOB;
