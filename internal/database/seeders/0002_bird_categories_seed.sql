-- Seed: kategori burung (general lookup).
-- Idempotent: aman dijalankan berulang kali.
-- ON CONFLICT DO UPDATE memastikan data existing ikut sinkron saat icon/name diperbarui.
INSERT INTO bird_categories (name, slug, icon) VALUES
    ('Murai Batu', 'murai_batu', 'murai_batu.svg'),
    ('Cucak Hijau', 'cucak_hijau', 'cucak_hijau.svg'),
    ('Kacer', 'kacer', 'kacer.svg'),
    ('Anis Merah','anis_merah', 'anis_merah.svg'),
    ('Kenari', 'kenari', 'kenari.svg'),
    ('Lovebird', 'lovebird', 'lovebird.svg')
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    icon = EXCLUDED.icon;
