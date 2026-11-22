-- Seed: akun demo untuk pengembangan.
-- Idempotent: aman dijalankan berulang kali.
INSERT INTO users (name, identifier, password_hash)
VALUES ('Kicau Mania', 'test@gmail.com', '$2a$10$GZo.EWIo5w6UVxx0YubDouDd48/Itn/qeDUg3nUlq27/Im2WWi8Hq')
ON CONFLICT (identifier) DO NOTHING;
