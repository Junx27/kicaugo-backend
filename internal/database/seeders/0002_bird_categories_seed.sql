-- Seed: kategori burung (general lookup).
-- Idempotent: aman dijalankan berulang kali.
-- ON CONFLICT DO UPDATE memastikan data existing ikut sinkron saat icon/name diperbarui.
INSERT INTO bird_categories (name, slug, icon) VALUES
    ('Murai Batu', 'murai_batu', 'assets/bird_categories/murai_batu.svg'),
    ('Cucak Hijau', 'cucak_hijau', 'assets/bird_categories/cucak_hijau.svg'),
    ('Kacer', 'kacer', 'assets/bird_categories/kacer.svg'),
    ('Anis Merah', 'anis_merah', 'aassets/bird_categories/ssets/bird_categories/anis_merah.svg'),
    ('Kenari', 'kenari', 'assets/bird_categories/kenari.svg'),
    ('Lovebird', 'lovebird', 'assets/bird_categories/lovebird.svg')
ON CONFLICT (slug) DO UPDATE
SET name = EXCLUDED.name,
    icon = EXCLUDED.icon;
