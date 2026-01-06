-- Revert: Rename email back to username
ALTER TABLE users RENAME COLUMN email TO username;

-- Revert: Rename index back
DROP INDEX IF EXISTS idx_users_email;
CREATE INDEX idx_users_username ON users(username);
