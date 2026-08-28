-- +goose Up
CREATE TYPE user_role AS ENUM ('admin');

ALTER TABLE admins RENAME TO users;
ALTER INDEX uq_admins_email RENAME TO uq_users_email;

ALTER TABLE users
    ADD COLUMN role user_role NOT NULL DEFAULT 'admin';

-- +goose Down
ALTER TABLE users DROP COLUMN role;
ALTER INDEX uq_users_email RENAME TO uq_admins_email;
ALTER TABLE users RENAME TO admins;
DROP TYPE user_role;
