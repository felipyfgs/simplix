-- +goose Up
-- +goose StatementBegin
-- Default admin user (password: admin123 — troque em produção)
INSERT INTO users (name, email, password_hash, role, availability)
VALUES (
    'Admin',
    'admin@simplix.local',
    '$2a$10$PMFwiSJ4gEFSHOEJYco9HeUTo93eGkQLvmFljsclZOArJrXYrg5i2',
    'admin',
    'online'
) ON CONFLICT (email) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM users WHERE email = 'admin@simplix.local';
-- +goose StatementEnd
