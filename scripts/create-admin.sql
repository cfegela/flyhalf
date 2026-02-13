-- Create an initial admin user
-- Default password: admin123 (change this immediately after first login!)
-- Password hash generated with bcrypt cost 12
-- nosemgrep: detected-bcrypt-hash

INSERT INTO users (email, password_hash, role, first_name, last_name, is_active)
VALUES (
    'admin@flyhalf.app',
    '$2a$12$R2iQS4ZXc0z1h7Oq2wAOKeqslDynZTXBkt9chHBIVIRUuUVO.nbPi',
    'admin',
    'System',
    'Administrator',
    true
)
ON CONFLICT (email) DO NOTHING;
