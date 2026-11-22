-- Seed: kategori burung (general lookup).
-- Idempotent: aman dijalankan berulang kali.
INSERT INTO bird_categories (name, slug, icon) VALUES
    ('Murai Batu', 'murai_batu', 'murai_batu'),
    ('Cucak Hijau', 'cucak_hijau', 'cucak_hijau'),
    ('Kacer', 'kacer', 'kacer'),
    ('Anis Merah', 'anis_merah', 'anis_merah'),
    ('Kenari', 'kenari', 'kenari'),
    ('Lovebird', 'lovebird', 'lovebird')
ON CONFLICT (slug) DO NOTHING;
