-- Rename username column to email
ALTER TABLE users RENAME COLUMN username TO email;

-- Rename index
DROP INDEX IF EXISTS idx_users_username;
CREATE INDEX idx_users_email ON users(email);
